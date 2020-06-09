package syncp_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/syncp"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	if os.Getenv("TEST_GIT") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	repo := "git@gitlab.com:gabrielha/keys-pub-git-test.git"
	program, err := syncp.NewGit(repo)
	require.NoError(t, err)

	rt := newTestRuntime(t)

	// Sync
	err = program.Sync(cfg, syncp.WithRuntime(rt))
	require.NoError(t, err)

	// Sync with new files
	testProgramSync(t, program, cfg, rt)

	// Sync again
	err = program.Sync(cfg, syncp.WithRuntime(rt))
	require.NoError(t, err)

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}

func TestGitFixtures(t *testing.T) {
	if os.Getenv("TEST_GIT") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	repo := "git@gitlab.com:gabrielha/keys-pub-git-test.git"
	program, err := syncp.NewGit(repo)
	require.NoError(t, err)

	testFixtures(t, program, cfg)
}
