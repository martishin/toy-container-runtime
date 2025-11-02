package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/martishin/toy-container-runtime/internal/build"
	"github.com/martishin/toy-container-runtime/internal/container"
)

func main() {
	root := &cobra.Command{
		Use:           "mini_container",
		SilenceUsage:  true,
		SilenceErrors: false,
		Short:         "A toy container runtime (namespaces, chroot, cgroups)",
	}

	// -------------------------------------------------------------------------
	// child (internal) — mini_container child <rootfs> <cmd> [args...]
	// -------------------------------------------------------------------------
	childCmd := &cobra.Command{
		Use:    "child <rootfs> <cmd> [args...]",
		Short:  "internal: run as the child process in new namespaces",
		Hidden: true,
		Args:   cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootfs := args[0]
			argv := args[1:]
			if err := container.RunChild(rootfs, argv); err != nil {
				return fmt.Errorf("child: %w", err)
			}
			return nil
		},
	}

	// -------------------------------------------------------------------------
	// build — mini_container build -f Dockerfile -t TAG -o ./rootfs [-C context]
	// -------------------------------------------------------------------------
	var buildFlags struct {
		df  string
		tag string
		out string
		ctx string
	}
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build rootfs from a Dockerfile (via docker build + export)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := build.Options{
				ContextDir:   buildFlags.ctx,
				Dockerfile:   buildFlags.df,
				ImageTag:     buildFlags.tag,
				OutputRootfs: buildFlags.out,
			}
			if err := build.BuildRootfsWithDocker(opts); err != nil {
				return fmt.Errorf("build: %w", err)
			}
			return nil
		},
	}
	buildCmd.Flags().StringVarP(&buildFlags.df, "file", "f", "Dockerfile", "Dockerfile path")
	buildCmd.Flags().StringVarP(&buildFlags.tag, "tag", "t", "mini/rootfs:latest", "image tag")
	buildCmd.Flags().StringVarP(&buildFlags.out, "output", "o", "./rootfs", "output rootfs dir")
	buildCmd.Flags().StringVarP(&buildFlags.ctx, "context", "C", ".", "build context dir")

	// -------------------------------------------------------------------------
	// run-rootfs — mini_container run-rootfs <rootfs> [--] <cmd> [args...]
	// -------------------------------------------------------------------------
	runRootfsCmd := &cobra.Command{
		Use:   "run-rootfs <rootfs> [--] <cmd> [args...]",
		Short: "Run a command inside the provided rootfs",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Accept both with and without "--"
			dash := cmd.ArgsLenAtDash()
			var rootfs string
			var argv []string

			if dash >= 0 {
				if len(args) < 1 || dash < 1 || len(args[dash:]) == 0 {
					return fmt.Errorf("usage: %s", cmd.UseLine())
				}
				rootfs = args[0]
				argv = args[dash:]
			} else {
				// no "--": treat args[0] as rootfs, the rest as command
				if len(args) < 2 {
					return fmt.Errorf("usage: %s", cmd.UseLine())
				}
				rootfs = args[0]
				argv = args[1:]
			}

			if err := container.Run(rootfs, argv); err != nil {
				return fmt.Errorf("run-rootfs: %w", err)
			}
			return nil
		},
	}

	// -------------------------------------------------------------------------
	// run — mini_container run -f Dockerfile [-C context] [--] <cmd> [args...]
	//        builds to ./rootfs, then runs the command
	// -------------------------------------------------------------------------
	var runFlags struct {
		df  string
		ctx string
	}
	runCmd := &cobra.Command{
		Use:   "run -f Dockerfile [-C context] [--] <cmd> [args...]",
		Short: "Build a rootfs from Dockerfile and run a command inside it",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Any remaining args after flags are the command (cobra strips "--")
			argv := args
			if len(argv) == 0 {
				return fmt.Errorf("usage: %s", cmd.UseLine())
			}

			// Build into ./rootfs with a temp tag
			opts := build.Options{
				ContextDir:   runFlags.ctx,
				Dockerfile:   runFlags.df,
				ImageTag:     "mini/rootfs:tmp",
				OutputRootfs: "./rootfs",
			}
			if err := build.BuildRootfsWithDocker(opts); err != nil {
				return fmt.Errorf("build: %w", err)
			}

			if err := container.Run("./rootfs", argv); err != nil {
				return fmt.Errorf("run: %w", err)
			}
			return nil
		},
	}
	runCmd.Flags().StringVarP(&runFlags.df, "file", "f", "Dockerfile", "Dockerfile path")
	runCmd.Flags().StringVarP(&runFlags.ctx, "context", "C", ".", "build context dir")

	// Wire up
	root.AddCommand(buildCmd, runRootfsCmd, runCmd, childCmd)

	// Execute
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
