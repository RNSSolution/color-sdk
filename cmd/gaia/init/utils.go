package init

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	amino "github.com/tendermint/go-amino"
	cfg "github.com/ColorPlatform/prism/config"
	"github.com/ColorPlatform/prism/crypto"
	"github.com/ColorPlatform/prism/libs/common"
	"github.com/ColorPlatform/prism/p2p"
	"github.com/ColorPlatform/prism/privval"
	"github.com/ColorPlatform/prism/types"

	"github.com/ColorPlatform/color-sdk/cmd/gaia/app"
	"github.com/ColorPlatform/color-sdk/codec"
	"github.com/ColorPlatform/color-sdk/server"
)

// ExportGenesisFile creates and writes the genesis configuration to disk. An
// error is returned if building or writing the configuration to file fails.
func ExportGenesisFile(
	genFile, chainID string, validators []types.GenesisValidator, appState json.RawMessage,
) error {

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		Validators: validators,
		AppState:   appState,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genFile)
}

// ExportGenesisFileWithTime creates and writes the genesis configuration to disk.
// An error is returned if building or writing the configuration to file fails.
func ExportGenesisFileWithTime(
	genFile, chainID string, validators []types.GenesisValidator,
	appState json.RawMessage, genTime time.Time,
) error {

	genDoc := types.GenesisDoc{
		GenesisTime: genTime,
		ChainID:     chainID,
		Validators:  validators,
		AppState:    appState,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genFile)
}

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(
	config *cfg.Config) (nodeID string, valPubKey crypto.PubKey, err error,
) {

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nodeID, valPubKey, err
	}

	nodeID = string(nodeKey.ID())
	server.UpgradeOldPrivValFile(config)

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := common.EnsureDir(filepath.Dir(pvKeyFile), 0777); err != nil {
		return nodeID, valPubKey, nil
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := common.EnsureDir(filepath.Dir(pvStateFile), 0777); err != nil {
		return nodeID, valPubKey, nil
	}

	valPubKey = privval.LoadOrGenFilePV(pvKeyFile, pvStateFile).GetPubKey()

	return nodeID, valPubKey, nil
}

// LoadGenesisDoc reads and unmarshals GenesisDoc from the given file.
func LoadGenesisDoc(cdc *amino.Codec, genFile string) (genDoc types.GenesisDoc, err error) {
	genContents, err := ioutil.ReadFile(genFile)
	if err != nil {
		return genDoc, err
	}

	if err := cdc.UnmarshalJSON(genContents, &genDoc); err != nil {
		return genDoc, err
	}

	return genDoc, err
}

func initializeEmptyGenesis(
	cdc *codec.Codec, genFile, chainID string, overwrite bool,
) (appState json.RawMessage, err error) {

	if !overwrite && common.FileExists(genFile) {
		return nil, fmt.Errorf("genesis.json file already exists: %v", genFile)
	}

	return codec.MarshalJSONIndent(cdc, app.NewDefaultGenesisState())
}
