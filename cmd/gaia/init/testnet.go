package init

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ColorPlatform/color-sdk/client/keys"

	"github.com/ColorPlatform/color-sdk/client"
	"github.com/ColorPlatform/color-sdk/cmd/gaia/app"
	"github.com/ColorPlatform/color-sdk/codec"
	srvconfig "github.com/ColorPlatform/color-sdk/server/config"
	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/auth"
	authtx "github.com/ColorPlatform/color-sdk/x/auth/client/txbuilder"
	"github.com/ColorPlatform/color-sdk/x/staking"
	tmconfig "github.com/ColorPlatform/prism/config"
	"github.com/ColorPlatform/prism/crypto"
	cmn "github.com/ColorPlatform/prism/libs/common"
	"github.com/ColorPlatform/prism/p2p"
	"github.com/ColorPlatform/prism/privval"
	"github.com/ColorPlatform/prism/types"
	tmtime "github.com/ColorPlatform/prism/types/time"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ColorPlatform/color-sdk/server"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagNodeCliHome       = "node-cli-home"
	flagStartingIPAddress = "starting-ip-address"
	flagNumLeagues        = "l"
)

const nodeDirPerm = 0755

// get cmd to initialize all files for tendermint testnet and application
func TestnetFilesCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Colord testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	gaiad testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			return initTestnet(config, cdc)
		},
	}

	cmd.Flags().Int(flagNumLeagues, 3,
		"Number of leagues to initialize the testnet with",
	)
	cmd.Flags().Int(flagNumValidators, 3,
		"Number of validators to initialize the testnet with",
	)
	cmd.Flags().StringP(flagOutputDir, "o", ".",
		"Directory to store initialization data for the testnet",
	)
	cmd.Flags().String(flagNodeDirPrefix, "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)",
	)
	cmd.Flags().String(flagNodeDaemonHome, "colord",
		"Home directory of the node's daemon configuration",
	)
	cmd.Flags().String(flagNodeCliHome, "colorcli",
		"Home directory of the node's cli configuration",
	)
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")

	cmd.Flags().String(
		client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created",
	)
	cmd.Flags().String(
		server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		"Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01uclr,0.001stake)",
	)

	return cmd
}

type nodeInfo struct {
	league      int
	nodeId      int
	nodeDir     string
	isValidator bool
}

var nodes []nodeInfo

func initTestnet(config *tmconfig.Config, cdc *codec.Codec) error {
	var chainID string

	outDir := viper.GetString(flagOutputDir)
	numValidators := viper.GetInt(flagNumValidators)
	numLeagues := viper.GetInt(flagNumLeagues)

	chainID = viper.GetString(client.FlagChainID)
	if chainID == "" {
		chainID = "colors-test-01"
	}
	genVals := make([]types.GenesisValidator, numValidators*numLeagues)
	monikers := make([]string, numLeagues*numValidators)
	nodeIDs := make([]string, numLeagues*numValidators)
	nodes := make([]nodeInfo, numLeagues*numValidators)
	valPubKeys := make([]crypto.PubKey, numLeagues*numValidators)

	gaiaConfig := srvconfig.DefaultConfig()
	gaiaConfig.MinGasPrices = viper.GetString(server.FlagMinGasPrices)

	config.P2P.MaxNumOutboundPeers = numLeagues*numValidators + 5
	config.P2P.MaxNumInboundPeers = numLeagues*numValidators + 15

	config.LogLevel = "main:info,state:info,*:error"
	config.ProfListenAddress = "0.0.0.0:6060"
	config.Consensus.TimeoutCommit = 200 * time.Millisecond

	var (
		accs     []app.GenesisAccount
		genFiles []string
	)

	// generate private keys, node IDs, and initial transactions
	for l := 0; l < numLeagues; l++ {
		for i := 0; i < numValidators; i++ {
			id := l*numValidators + i
			nodeDirName := fmt.Sprintf("%s%d", viper.GetString(flagNodeDirPrefix), id)
			nodeDaemonHomeName := viper.GetString(flagNodeDaemonHome)
			nodeCliHomeName := viper.GetString(flagNodeCliHome)
			nodeDir := filepath.Join(outDir, nodeDirName, nodeDaemonHomeName)
			clientDir := filepath.Join(outDir, nodeDirName, nodeCliHomeName)
			gentxsDir := filepath.Join(outDir, "gentxs")

			node := nodeInfo{l, id, nodeDir, true}
			nodes[id] = node

			if id == 0 {
				config.LogLevel = "p2p:debug,consensus:debug,main:info,state:info,*:error"
			} else {
				config.LogLevel = "main:info,state:info,*:error"
			}

			config.SetRoot(nodeDir)

			err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			err = os.MkdirAll(clientDir, nodeDirPerm)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			monikers = append(monikers, nodeDirName)
			config.Moniker = nodeDirName

			ip, err := getIP(id, viper.GetString(flagStartingIPAddress))
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			nodeIDs[i], valPubKeys[i], err = InitializeNodeValidatorFiles(config)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
			genFiles = append(genFiles, config.GenesisFile())

			// buf := client.BufferStdin()
			// prompt := fmt.Sprintf(
			// 	"Password for account '%s' (default %s):", nodeDirName, app.DefaultKeyPass,
			// )

			// keyPass, err := client.GetPassword(prompt, buf)
			// if err != nil && keyPass != "" {
			// 	// An error was returned that either failed to read the password from
			// 	// STDIN or the given password is not empty but failed to meet minimum
			// 	// length requirements.
			// 	return err
			// }

			// if keyPass == "" {
			keyPass := app.DefaultKeyPass
			// }

			addr, secret, err := server.GenerateSaveCoinKey(clientDir, nodeDirName, keyPass, true)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			info := map[string]string{"secret": secret}

			cliPrint, err := json.Marshal(info)
			if err != nil {
				return err
			}

			// save private key seed words
			err = writeFile(fmt.Sprintf("%v.json", "key_seed"), clientDir, cliPrint)
			if err != nil {
				return err
			}

			accStakingTokens := sdk.TokensFromTendermintPower(17777778)
			accs = append(accs, app.GenesisAccount{
				Address: addr,
				Coins: sdk.Coins{
					sdk.NewCoin(sdk.DefaultBondDenom, accStakingTokens),
				},
			})

			valTokens := sdk.TokensFromTendermintPower(45000)
			msg := staking.NewMsgCreateValidator(
				sdk.ValAddress(addr),
				valPubKeys[i],
				sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
				staking.NewDescription(nodeDirName, "", "", ""),
				staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
				sdk.OneInt(), sdk.NumberToInt(l), sdk.NumberToInt(i),
			)
			kb, err := keys.NewKeyBaseFromDir(clientDir)
			if err != nil {
				return err
			}
			tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, memo)
			txBldr := authtx.NewTxBuilderFromCLI().WithChainID(chainID).WithMemo(memo).WithKeybase(kb)

			signedTx, err := txBldr.SignStdTx(nodeDirName, app.DefaultKeyPass, tx, false)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			txBytes, err := cdc.MarshalJSON(signedTx)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			// gather gentxs folder
			err = writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBytes)
			if err != nil {
				_ = os.RemoveAll(outDir)
				return err
			}

			gaiaConfigFilePath := filepath.Join(nodeDir, "config/colord.toml")
			srvconfig.WriteConfigFile(gaiaConfigFilePath, gaiaConfig)

			pvKeyFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorKey)
			pvStateFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorState)

			pv := privval.LoadFilePV(pvKeyFile, pvStateFile)
			genVals[l*numValidators+i] = types.GenesisValidator{
				League:  l,
				NodeId:  l*numValidators + i,
				Address: pv.GetPubKey().Address(),
				PubKey:  pv.GetPubKey(),
				Power:   1,
				Name:    nodeDirName,
			}

		}
	}

	if err := initGenFiles(cdc, chainID, accs, genFiles, numLeagues*numValidators, genVals); err != nil {
		return err
	}

	err := collectGenFiles(
		cdc, config, chainID, genVals, monikers, nodeIDs, valPubKeys, numLeagues*numValidators,
		outDir, viper.GetString(flagNodeDirPrefix), viper.GetString(flagNodeDaemonHome), nodes,
	)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully initialized %d node directories\n", numValidators*numLeagues)
	fmt.Printf("Default password is := 12345678 \n")

	// Create league docs and fill it
	leaguesDoc := &types.LeaguesDoc{
		Leagues: numLeagues,
		Peers:   fillLeagues(nodes, config),
	}

	for _, node := range nodes {
		if err := leaguesDoc.SaveAs(filepath.Join(node.nodeDir, config.BaseConfig.Leagues)); err != nil {
			_ = os.RemoveAll(outDir)
			return err
		}
	}

	return nil
}

// function to fill leagues with nodes
func fillLeagues(nodes []nodeInfo, config *tmconfig.Config) []types.LeaguePeer {
	result := make([]types.LeaguePeer, len(nodes))
	for i := range result {
		node := nodes[i]
		result[i] = types.LeaguePeer{
			League:   node.league,
			NodeId:   node.nodeId,
			PubKey:   readPubKey(node.nodeDir, config),
			Hostname: "node" + strconv.FormatInt(int64(i), 10),
		}
	}
	return result
}

// function to read public keys of nodes (to add in leagues.json file)
func readPubKey(cfgDir string, config *tmconfig.Config) crypto.PubKey {
	config.SetRoot(cfgDir)
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(fmt.Sprintf("Failed to read node key: %s", config.NodeKeyFile()))
	}
	return nodeKey.PubKey()
}

func initGenFiles(
	cdc *codec.Codec, chainID string, accs []app.GenesisAccount,
	genFiles []string, totalnumValidators int, genVals []types.GenesisValidator,
) error {

	appGenState := app.NewDefaultGenesisState()
	appGenState.Accounts = accs

	appGenStateJSON, err := codec.MarshalJSONIndent(cdc, appGenState)
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: genVals,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < totalnumValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func collectGenFiles(
	cdc *codec.Codec, config *tmconfig.Config, chainID string, genvals []types.GenesisValidator,
	monikers, nodeIDs []string, valPubKeys []crypto.PubKey,
	totalnumValidators int, outDir, nodeDirPrefix, nodeDaemonHomeName string, nodes []nodeInfo,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < totalnumValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outDir, nodeDirName, nodeDaemonHomeName)
		gentxsDir := filepath.Join(outDir, "gentxs")
		moniker := monikers[i]
		config.Moniker = nodeDirName
		config.Consensus.UseLeagues = true
		config.Consensus.League = nodes[i].league
		config.Consensus.NodeId = i
		config.Consensus.CreateEmptyBlocksInterval = 0 * time.Second
		config.P2P.AddrBookStrict = false
		config.P2P.AllowDuplicateIP = true
		config.P2P.MaxNumInboundPeers = 100
		//config.P2P.MaxNumOutboundPeers = 100

		config.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := newInitConfig(chainID, gentxsDir, moniker, nodeID, valPubKey)

		genDoc, err := LoadGenesisDoc(cdc, config.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genAppStateFromConfig(cdc, config, initCfg, genDoc)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := config.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		err = ExportGenesisFileWithTime(genFile, chainID, genvals, appState, genTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (string, error) {
	var (
		ip  string
		err error
	)

	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
	} else {
		ip, err = calculateIP(startingIPAddr, i)
		if err != nil {
			return "", err
		}
	}

	return ip, nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := cmn.EnsureDir(writePath, 0700)
	if err != nil {
		return err
	}

	err = cmn.WriteFile(file, contents, 0600)
	if err != nil {
		return err
	}

	return nil
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}
