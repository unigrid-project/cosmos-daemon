package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/unigrid-project/pax/app"
)

const DefaultHedgehogUrl = "https://149.102.147.45:39886"

// testing localhost hedgehog
//const DefaultHedgehogUrl = "https://127.0.0.1:39886"

// NewRootCmd creates a new root command for paxd. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	initSDKConfig()

	var (
		txConfigOpts       tx.ConfigOptions
		autoCliOpts        autocli.AppOptions
		moduleBasicManager module.BasicManager
		clientCtx          client.Context
	)

	//if err := depinject.InjectDebug(
	//depinject.FileVisualizer("/home/evan/work/cosmos-daemon/output.dot"),
	if err := depinject.Inject(
		depinject.Configs(
			app.AppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
			),
			depinject.Provide(
				ProvideClientContext,
				ProvideKeyring,
			),
		),
		&txConfigOpts,
		&autoCliOpts,
		&moduleBasicManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}

	// Since the IBC modules don't support dependency injection, we need to
	// manually add the modules to the basic manager on the client side.
	// This needs to be removed after IBC supports App Wiring.
	app.AddIBCModuleManager(moduleBasicManager)

	rootCmd := &cobra.Command{
		Use:           app.Name + "d",
		Short:         "Start pax node",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx = clientCtx.WithCmdContext(cmd.Context())
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = config.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL.
			txConfigOpts.EnabledSignModes = append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts.TextualCoinMetadataQueryFn = txmodule.NewGRPCCoinMetadataQueryFn(clientCtx)
			txConfigWithTextual, err := tx.NewTxConfigWithOptions(
				codec.NewProtoCodec(clientCtx.InterfaceRegistry),
				txConfigOpts,
			)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithTxConfig(txConfigWithTextual)
			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	// param for the hedgehog url to be passed at startup
	rootCmd.PersistentFlags().StringVar(&HedgehogUrl, "hedgehog", "", "Pass the Hedgehog URL")
	//fmt.Println("Value of --hedgehog flag:", HedgehogUrl)

	viper.BindPFlag("hedgehog.hedgehog_url", rootCmd.PersistentFlags().Lookup("hedgehog"))

	initRootCmd(rootCmd, clientCtx.TxConfig, clientCtx.InterfaceRegistry, clientCtx.Codec, moduleBasicManager)

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        strings.ReplaceAll(app.Name, "-", ""),
		flags.FlagKeyringBackend: "test",
	})

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			f.Value.Set(val)
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfig client.TxConfig,
	legacyAmino *codec.LegacyAmino,
) client.Context {
	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.Name) // env variable prefix

	// Read the config again to overwrite the default values with the values from the config file
	clientCtx, _ = config.ReadFromClientConfig(clientCtx)

	return clientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return keyring.NewAutoCLIKeyring(kb)
}

// InitializeConfig reads the hedgehog configuration file and ENV variables if set.
// if there is no flag or config we use DefaultHedgehogUrl
func InitializeConfig(home string) {
	viper.SetConfigName("hedgehog")
	viper.SetConfigType("toml")
	configPath := filepath.Join(home, "config")
	viper.AddConfigPath(configPath)

	// Print the full path of the file it's looking for
	fullPath := filepath.Join(configPath, "hedgehog.toml")
	fmt.Println("Searching for config file at:", fullPath)

	err := viper.ReadInConfig()
	if err != nil {
		// Check if the error is due to the file not being found
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore or handle as desired
			fmt.Println("Warning: Hedgehog Config file not found. Using default settings.")
		} else {
			// Some other error occurred; handle or print it
			fmt.Printf("Error reading config file: %s", err)
			return
		}
	}

	// If HedgehogUrl flag value is set, prioritize it over the configuration file value
	if HedgehogUrl != "" {
		viper.Set("hedgehog.hedgehog_url", HedgehogUrl)
	}

	// Get the value of hedgehog_url
	hedgehogURL := viper.GetString("hedgehog.hedgehog_url")

	// If hedgehogURL is empty, set it to the default value
	if hedgehogURL == "" {
		hedgehogURL = DefaultHedgehogUrl
		viper.Set("hedgehog.hedgehog_url", DefaultHedgehogUrl) // Set the default value in viper
	}

	fmt.Println("Hedgehog URL:", viper.GetString("hedgehog.hedgehog_url"))
}
