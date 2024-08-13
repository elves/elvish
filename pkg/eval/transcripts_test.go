package eval_test

import (
	"embed"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/transcript"
)

var (
	//go:embed *.elvts *.elv
	transcripts     embed.FS
	transcriptNodes = must.OK1(transcript.ParseFromFS(transcripts))
	transcriptCodes = extractAllCodes(transcriptNodes)
)

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptNodes(t, transcriptNodes,
		"args", func(ev *eval.Evaler, arg string) {
			ev.Args = vals.MakeListSlice(strings.Fields(arg))
		},
		"recv-bg-job-notification-in-global", func(ev *eval.Evaler) {
			noteCh := make(chan string, 10)
			ev.BgJobNotify = func(s string) { noteCh <- s }
			ev.ExtendGlobal(eval.BuildNs().
				AddGoFn("recv-bg-job-notification", func() any { return <-noteCh }))
		},
		"with-temp-home", func(t *testing.T) { testutil.TempHome(t) },
		"reseed-afterwards", func(t *testing.T) {
			t.Cleanup(func() {
				//lint:ignore SA1019 Reseed to make other RNG-dependent tests non-deterministic
				rand.Seed(time.Now().UTC().UnixNano())
			})
		},
		"check-exit-code-afterwards", func(t *testing.T, arg string) {
			var exitCodes []int
			testutil.Set(t, eval.OSExit, func(i int) {
				exitCodes = append(exitCodes, i)
			})
			wantExitCode := must.OK1(strconv.Atoi(arg))
			t.Cleanup(func() {
				if len(exitCodes) != 1 {
					t.Errorf("os.Exit called %d times, want once", len(exitCodes))
				} else if exitCodes[0] != wantExitCode {
					t.Errorf("os.Exit called with %d, want %d", exitCodes[0], wantExitCode)
				}
			})
		},
		"check-pre-exit-hook-afterwards", func(t *testing.T, ev *eval.Evaler) {
			testutil.Set(t, eval.OSExit, func(int) {})
			calls := 0
			ev.PreExitHooks = append(ev.PreExitHooks, func() { calls++ })
			t.Cleanup(func() {
				if calls != 1 {
					t.Errorf("pre-exit hook called %v times, want 1", calls)
				}
			})
		},
		"add-bad-var", func(ev *eval.Evaler, arg string) {
			name, allowedSetsString, _ := strings.Cut(arg, " ")
			allowedSets := must.OK1(strconv.Atoi(allowedSetsString))
			ev.ExtendGlobal(eval.BuildNs().AddVar(name, &badVar{allowedSets}))
		},
		"tmp-lib-dir", func(t *testing.T, ev *eval.Evaler) {
			libdir := testutil.TempDir(t)
			ev.LibDirs = []string{libdir}
			ev.ExtendGlobal(eval.BuildNs().
				AddVar("lib", vars.NewReadOnly(libdir)))
		},
		"two-tmp-lib-dirs", func(t *testing.T, ev *eval.Evaler) {
			libdir1 := testutil.TempDir(t)
			libdir2 := testutil.TempDir(t)
			ev.LibDirs = []string{libdir1, libdir2}
			ev.ExtendGlobal(eval.BuildNs().
				AddVar("lib1", vars.NewReadOnly(libdir1)).
				AddVar("lib2", vars.NewReadOnly(libdir2)))
		},
		"add-var-in-builtin", func(ev *eval.Evaler) {
			addVar := func(name string, val any) {
				ev.ExtendGlobal(eval.BuildNs().AddVar(name, vars.FromInit(val)))
			}
			ev.ExtendBuiltin(eval.BuildNs().AddGoFn("add-var", addVar))
		},
		"test-time-scale-in-global", func(ev *eval.Evaler) {
			ev.ExtendGlobal(eval.BuildNs().
				AddVar("test-time-scale", vars.NewReadOnly(testutil.TestTimeScale())))
		},
		"mock-get-home-error", func(t *testing.T, msg string) {
			err := errors.New(msg)
			testutil.Set(t, eval.GetHome,
				func(name string) (string, error) { return "", err })
		},
		"force-eval-source-count", func(t *testing.T, arg string) {
			c := must.OK1(strconv.Atoi(arg))
			testutil.Set(t, eval.NextEvalCount, func() int { return c })
		},
		"mock-time-after", func(t *testing.T) {
			testutil.Set(t, eval.TimeAfter,
				func(fm *eval.Frame, d time.Duration) <-chan time.Time {
					fmt.Fprintf(fm.ByteOutput(), "slept for %s\n", d)
					return time.After(0)
				})
		},
		"mock-benchmark-run-durations", func(t *testing.T, arg string) {
			// The benchmark command calls time.Now once before a run and once
			// after a run.
			var ticks []int64
			for i, field := range strings.Fields(arg) {
				d := must.OK1(strconv.ParseInt(field, 0, 64))
				if i == 0 {
					ticks = append(ticks, 0, d)
				} else {
					last := ticks[len(ticks)-1]
					ticks = append(ticks, last, last+d)
				}
			}
			testutil.Set(t, eval.TimeNow, func() time.Time {
				if len(ticks) == 0 {
					panic("mock TimeNow called more than len(ticks)")
				}
				v := ticks[0]
				ticks = ticks[1:]
				return time.Unix(v, 0)
			})
		},
		"inject-time-after-with-sigint-or-skip", injectTimeAfterWithSIGINTOrSkip,
		"mock-getwd-error", func(t *testing.T, msg string) {
			err := errors.New(msg)
			testutil.Set(t, eval.Getwd, func() (string, error) { return "", err })
		},
		"mock-no-other-home", func(t *testing.T) {
			testutil.Set(t, eval.GetHome, func(name string) (string, error) {
				switch name {
				case "":
					return fsutil.GetHome("")
				default:
					return "", fmt.Errorf("don't know home of %v", name)
				}
			})
		},
		"mock-one-other-home", func(t *testing.T, ev *eval.Evaler) {
			otherHome := testutil.TempDir(t)
			ev.ExtendGlobal(eval.BuildNs().AddVar("other-home", vars.NewReadOnly(otherHome)))
			testutil.Set(t, eval.GetHome, func(name string) (string, error) {
				switch name {
				case "":
					return fsutil.GetHome("")
				case "other":
					return otherHome, nil
				default:
					return "", fmt.Errorf("don't know home of %v", name)
				}
			})
		},
		"go-fns-mod-in-global", func(ev *eval.Evaler) {
			ev.ExtendGlobal(eval.BuildNs().AddNs("go-fns", goFnsMod))
		},
		"call-hook-in-global", func(ev *eval.Evaler) {
			callHook := func(fm *eval.Frame, name string, hook vals.List, args ...any) {
				evalCfg := &eval.EvalCfg{Ports: []*eval.Port{fm.Port(0), fm.Port(1), fm.Port(2)}}
				eval.CallHook(fm.Evaler, evalCfg, name, hook, args...)
			}
			ev.ExtendGlobal(eval.BuildNs().AddGoFn("call-hook", callHook))
		},
	)
}

var errBadVar = errors.New("bad var")

type badVar struct{ allowedSets int }

func (v *badVar) Get() any { return nil }

func (v *badVar) Set(any) error {
	if v.allowedSets == 0 {
		return errBadVar
	}
	v.allowedSets--
	return nil
}

func extractAllCodes(nodes []*transcript.Node) []string {
	var codes []string
	for _, node := range nodes {
		var codeBuf strings.Builder
		for i, interaction := range node.Interactions {
			if i > 0 {
				codeBuf.WriteByte('\n')
			}
			codeBuf.WriteString(interaction.Code)
		}
		codes = append(codes, codeBuf.String())
		codes = append(codes, extractAllCodes(node.Children)...)
	}
	return codes
}
