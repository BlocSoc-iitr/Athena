package backfill
// GetCommonFlags returns a set of common flags for CLI commands.
import (
	"github.com/urfave/cli/v2"
)
func GetCommonFlags() []cli.Flag {
    return []cli.Flag{
        &cli.StringFlag{
            Name:     "json_rpc",
            Usage:    "JSON RPC endpoint URL",
            Required: true,
        },
        &cli.StringFlag{
            Name:     "db_url",
            Usage:    "Database URL",
            Required: false,
        },
        &cli.IntFlag{
            Name:     "from_block",
            Usage:    "Start block number",
            Required: true,
        },
        &cli.IntFlag{
            Name:     "to_block",
            Usage:    "End block number",
            Required: true,
        },
        &cli.StringFlag{
            Name:     "block_file",
            Usage:    "Path to the block file",
            Required: false,
        },
        &cli.StringFlag{
            Name:     "transaction_file",
            Usage:    "Path to the transaction file",
            Required: false,
        },
        &cli.StringFlag{
            Name:     "event_file",
            Usage:    "Path to the event file",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "decode_abis",
            Usage:    "Decode ABIs",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "all_abis",
            Usage:    "Use all ABIs",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "no_interaction",
            Usage:    "Skip user interaction",
            Required: false,
        },
    }
}
func GetBackfillFlags(c *cli.Context) map[string]interface{} {
	// Initialize a map to store the flags
	flags := make(map[string]interface{})

	// Extract flags from the CLI context
	if c.IsSet("json_rpc") {
		flags["json_rpc"] = c.String("json_rpc")
	}
	if c.IsSet("db_url") {
		flags["db_url"] = c.String("db_url")
	}
	if c.IsSet("from_block") {
		flags["from_block"] = c.Int("from_block")
	}
	if c.IsSet("to_block") {
		flags["to_block"] = c.Int("to_block")
	}
	if c.IsSet("block_file") {
		flags["block_file"] = c.String("block_file")
	}
	if c.IsSet("transaction_file") {
		flags["transaction_file"] = c.String("transaction_file")
	}
	if c.IsSet("event_file") {
		flags["event_file"] = c.String("event_file")
	}
	if c.IsSet("decode_abis") {
		flags["decode_abis"] = c.Bool("decode_abis")
	}
	if c.IsSet("all_abis") {
		flags["all_abis"] = c.Bool("all_abis")
	}
	if c.IsSet("no_interaction") {
		flags["no_interaction"] = c.Bool("no_interaction")
	}

	// Add other flags as needed
	// Example: flags["some_flag"] = c.String("some_flag")

	return flags
}
func GetEventFlags() []cli.Flag {
    return []cli.Flag{
        &cli.StringFlag{
            Name:     "contract_address",
            Usage:    "Contract address for event backfill",
            Required: true,
        },
        &cli.StringFlag{
            Name:     "event_name",
            Usage:    "Name of the event",
            Required: false,
        },
        &cli.IntFlag{
            Name:     "batch_size",
            Usage:    "Batch size for processing events",
            Required: false,
        },
        &cli.StringFlag{
            Name:     "event_file",
            Usage:    "Path to the event file",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "decode_abis",
            Usage:    "Decode ABIs",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "all_abis",
            Usage:    "Use all ABIs",
            Required: false,
        },
        &cli.BoolFlag{
            Name:     "no_interaction",
            Usage:    "Skip user interaction",
            Required: false,
        },
    }
}