module github.com/ColorPlatform/color-sdk

require (
	bou.ke/monkey v1.0.1 // indirect
	github.com/ColorPlatform/prism v0.31.6-0.20191122022543-ce78f8f1af2e
	github.com/RNSSolution/iavl v0.0.1
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d
	github.com/cosmos/go-bip39 v0.0.0-20180618194314-52158e4697b8
	github.com/cosmos/ledger-cosmos-go v0.10.3
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.0
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.6
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/otiai10/copy v0.0.0-20180813032824-7e9a647135a1
	github.com/otiai10/curr v0.0.0-20150429015615-9b4961190c95 // indirect
	github.com/otiai10/mint v1.2.3 // indirect
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190227231451-bbced9601137 // indirect
	github.com/rakyll/statik v0.1.4
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/iavl v0.12.1
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/crypto v0.0.0-20190228161510-8dd112bcdc25
	google.golang.org/grpc v1.19.0 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

//replace github.com/ColorPlatform/prism => ../prism

go 1.13
