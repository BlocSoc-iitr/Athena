package backfill
import (
	"fmt"
	"bufio"
	"os"
	cli2 "github.com/urfave/cli/v2"
	"github.com/sirupsen/logrus"
	"github.com/DarkLord017/athena/athena/types"
	"github.com/DarkLord017/athena/athena/backfill"
	//"github.com/DarkLord017/athena/athena/decoder"
)
var logger = logrus.New()//logrus is a popular logging package in Go that provides structured logging with various formatting and output options.
func init(){//setup tasks
	logger.SetFormatter(&logrus.TextFormatter{ //formats log entries as plain text.
		DisableColors: false,
		FullTimestamp: true,//displays the full timestamp in the log entry.
	})
	logger.SetOutput(os.Stdout)//os.Stdout means that log messages will be written to the standard outpu
	logger.SetLevel(logrus.InfoLevel)
}
//Defining Subcommands:
func NewStarknetCommand() *cli2.Command {
	return &cli2.Command{
		Name:  "starknet",
		Usage: "Backfill StarkNet data from RPC",
		Subcommands: []*cli2.Command{
			{
				Name:   "full_blocks",
				Usage:  "Backfill StarkNet Blocks, Transactions, Receipts and Events",
				Action: full_blocks,
				Flags:  GetCommonFlags(),//decorators
			},
			{
				Name:   "transactions",
				Usage:  "Backfill StarkNet Transactions and Blocks",
				Action: transactions,
				Flags:  GetCommonFlags(),
			},
			{
				Name:   "events",
				Usage:  "Backfill & ABI Decode StarkNet Events for a Contract",
				Action: events,
				Flags:  GetEventFlags(),
			},
		},
	}
}
func full_blocks(c *cli2.Context) error {
    logrus.Info("Starting full blocks backfill...")
	// Retrieve flags from CLI context
	kwargs := GetBackfillFlags(c)

	backfillPlan, err := NewBackfillPlanFromCLI(//yet to be defined : similar as from_cli function is entro , will be defined in planner
		types.StarkNet,//(check all arguments ) //Network should be starknet
		types.FullBlocks,//backfill type should be all blocks 
		[]string{"json_rpc"},
		kwargs,
	)
	if err != nil {
		logger.Error("Failed to create backfill plan: ", err)
		return err
	}
	if !backfillPlan.NoInteraction {//to be defined in planner
		backfillPlan.PrintBackfillPlan(logrus.Info)//check
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Execute Backfill? [y/n] ")
        confirm, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		confirm = confirm[:len(confirm)-1]
		if confirm == "n" {
			return nil
		}
	}
	killer := backfill.New_Gracfull_Killer()
	backfillPlan.ExecuteBackfill(logrus.Info, killer)
	return nil
}
func transactions(c *cli2.Context) error {
	// Placeholder for transactions command implementation
	fmt.Println("Transactions backfill not yet implemented.")
	return nil
}
func events(c *cli2.Context) error {
    logrus.Info("Starting full blocks backfill...")
	kwargs := GetEventFlags(c)
	backfillPlan, err := NewBackfillPlanFromCLI(
		types.StarkNet,
		types.Events,
		[]string{"json_rpc"},
		kwargs,
	)
	if err != nil {
		logger.Error(err)
		return err
	}
	if !backfillPlan.NoInteraction {
		backfillPlan.PrintBackfillPlan(logrus.Info)
		reader := bufio.NewReader(os.Stdin)
        fmt.Print("Execute Backfill? [y/n] ")
        confirm, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if confirm == "n" {
			return nil
		}
	}
	killer := backfill.New_Gracfull_Killer()
	backfillPlan.ExecuteBackfill(logrus.Info, killer)
	return nil
}