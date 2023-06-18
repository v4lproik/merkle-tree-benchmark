package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/v4lproik/merkle-tree/pkg"
	"os"
	"os/signal"
	"syscall"
)

var buildCmd = &cobra.Command{
	Use:          "build",
	Short:        "build a merkle tree",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			mt  *pkg.MerkleTree
			err error
		)

		// initiate context
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		defer stop()

		// create conf
		hash := pkg.Hash(viper.GetString(projectName + ".hash"))
		if !hash.IsValid() {
			return err
		}
		var hashPool *pkg.HashPool
		if viper.GetBool(projectName + ".performance.reuse-buffer-allocation") {
			hashPool = pkg.NewHashPool(hash.Hash())
		}

		// fetch tree data
		_data := viper.GetStringSlice(projectName + ".data")
		data := make([]pkg.Data, len(_data))
		for i, d := range _data {
			data[i] = &pkg.StringData{
				Value: d,
			}
		}

		// use tree builder and build the tree
		if mt, err = pkg.NewMerkleTreeBuilder().
			WithHasher(&pkg.Hasher{
				IsSort: viper.GetBool(projectName + ".sort"),
				Hash:   hash,
				Pool:   hashPool,
			}).
			WithMaxGoroutine(viper.GetUint32(projectName+".performance.max-goroutine")).
			Build(ctx, data); err != nil {
			return err
		}

		// display merkle tree root
		log.Infof("merkle root hash: %x", mt.Root.Hash)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringSlice("data", []string{}, "data to insert into the merkle tree")
	_ = viper.BindPFlag(projectName+".data", buildCmd.Flag("data"))

	rootCmd.Flags().Uint("max-goroutine", 1000, "max goroutine")
	_ = viper.BindPFlag(projectName+".performance.max-goroutine", rootCmd.Flag("max goroutine"))

	rootCmd.Flags().Bool("reuse-buffer-allocation", true, "reuse buffer allocation")
	_ = viper.BindPFlag(projectName+".performance.reuse-buffer-allocation", rootCmd.Flag("reuse buffer allocation"))

	rootCmd.Flags().Bool("sort", true, "sort")
	_ = viper.BindPFlag(projectName+".sort", rootCmd.Flag("sort"))

}
