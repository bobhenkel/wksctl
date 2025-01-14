package disable

import (
	"errors"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/wksctl/cmd/wksctl/profile/constants"
	"github.com/weaveworks/wksctl/pkg/git"
)

var Cmd = &cobra.Command{
	Use:   "disable",
	Short: "disable profile",
	Long: `To disable the profile, run

wksctl profile disable --git-url=<profile_repository> [--revision=master] [--push=true]

Please make sure that there is no staged change on the current branch before disable a profile.
`,
	Args: profileDisableArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return profileDisable(profileDisableParams)
	},
	SilenceUsage: true,
}

type profileDisableFlags struct {
	gitUrl     string
	push       bool
	profileDir string
}

var profileDisableParams profileDisableFlags

func init() {
	Cmd.Flags().StringVar(&profileDisableParams.profileDir, "profile-dir", "profiles", "specify a directory for storing profiles")
	Cmd.Flags().StringVar(&profileDisableParams.gitUrl, "git-url", "", "enable profile from the Git URL")
	Cmd.Flags().BoolVar(&profileDisableParams.push, "push", true, "auto push after disable the profile")
}

func profileDisableArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("profile disable does not require any argument")
	}
	return nil
}

func profileDisable(params profileDisableFlags) error {
	repoUrl := params.gitUrl
	if repoUrl == constants.AppDevAlias {
		repoUrl = constants.AppDevRepoURL
	}

	if err := git.IsGitURL(repoUrl); err != nil {
		return err
	}

	hostName, repoName, err := git.HostAndRepoPath(repoUrl)
	if err != nil {
		return err
	}

	profilePath := path.Join(params.profileDir, hostName, repoName)
	// profilePath should exist
	if _, err := os.Stat(profilePath); err != nil {
		return err
	}

	// check if there is staged changes
	if err := git.HasNoStagedChanges(); err != nil {
		return err
	}

	log.Info("Removing profile from the local repository...")
	if err := git.RmRecursive(profilePath); err != nil {
		return err
	}
	log.Info("Removed profile from the local repository.")

	if err := git.Commit(fmt.Sprintf("Disable profile: %q", profilePath)); err != nil {
		return err
	}

	// Similar to enable, the default behaviour is auto-commit and push
	if params.push {
		if err := git.Push(); err != nil {
			return err
		}
	}

	return nil
}
