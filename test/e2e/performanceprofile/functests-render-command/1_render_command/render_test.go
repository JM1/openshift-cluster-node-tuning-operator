package __render_command_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	defaultExpectedDir      = "default"
	pinnedExpectedDir       = "pinned"
	bootstrapExpectedDir    = "bootstrap"
	noRefExpectedDir        = "no-ref"
	cpuFrequencyExpectedDir = defaultExpectedDir + "/" + "cpuFrequency"
	armExpectedDir          = defaultExpectedDir + "/" + "arm"
	ppNoRPSExpectedDir      = defaultExpectedDir + "/" + "pp-norps"
)

var (
	assetsOutDir                                            string
	assetsInDir                                             string
	assetsInDirs, assetsCpuFrequencyInDirs, assetsARMInDirs []string
	ppDir, ppDirNoRPS                                       string
	ppCpuFrequencyDir                                       string
	armDir                                                  string
	testDataPath                                            string
	defaultPinnedDir                                        string
	snoLegacyPinnedDir                                      string
	bootstrapPPDir                                          string
	extraMCPDir                                             string
)

var _ = Describe("render command e2e test", func() {

	BeforeEach(func() {
		assetsOutDir = createTempAssetsDir()
		assetsInDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "base", "performance")
		bootstrapPPDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "bootstrap-cluster", "performance")
		extraMCPDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "bootstrap-cluster", "extra-mcp")
		ppDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "manual-cluster", "performance")
		ppDirNoRPS = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "manual-cluster", "performance-norps")
		ppCpuFrequencyDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "manual-cluster", "cpuFrequency")
		armDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "manual-cluster", "arm")
		defaultPinnedDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "pinned-cluster", "default")
		snoLegacyPinnedDir = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "cluster-setup", "pinned-cluster", "single-node-legacy")
		testDataPath = filepath.Join(workspaceDir, "test", "e2e", "performanceprofile", "testdata")
		assetsInDirs = []string{assetsInDir, ppDir}
		assetsARMInDirs = []string{assetsInDir, armDir}
		assetsCpuFrequencyInDirs = []string{assetsInDir, ppCpuFrequencyDir}
	})

	Context("With a single performance-profile", func() {
		It("Gets cli args and produces the expected components to output directory", func() {

			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(assetsInDirs, ","),
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, defaultExpectedDir)

		})

		It("Gets environment variables and produces the expected components to output directory", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Env = append(cmd.Env,
				fmt.Sprintf("ASSET_INPUT_DIR=%s", strings.Join(assetsInDirs, ",")),
				fmt.Sprintf("ASSET_OUTPUT_DIR=%s", assetsOutDir),
			)
			runAndCompare(cmd, defaultExpectedDir)
		})

		It("Must fail to restore legacy and wrong legacy owner reference if uid is missing", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(assetsInDirs, ","),
				"--asset-output-dir", assetsOutDir,
				"--owner-ref", "k8s",
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			_, err := cmd.Output()
			Expect(err).To(HaveOccurred(), logStderr(err))
		})

		It("Must not set any owner reference if disabled explicitly", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(assetsInDirs, ","),
				"--asset-output-dir", assetsOutDir,
				"--owner-ref", "none",
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Env = append(cmd.Env,
				fmt.Sprintf("ASSET_INPUT_DIR=%s", strings.Join(assetsInDirs, ",")),
				fmt.Sprintf("ASSET_OUTPUT_DIR=%s", assetsOutDir),
			)
			runAndCompare(cmd, noRefExpectedDir)
		})

		It("Given cpu frequencies, should render appropriate sysfs configs", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(assetsCpuFrequencyInDirs, ","),
				"--asset-output-dir", assetsOutDir,
				"--owner-ref", "none",
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, cpuFrequencyExpectedDir)
		})

		It("should produces the expected components for ARM cluster", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(assetsARMInDirs, ","),
				"--asset-output-dir", assetsOutDir,
				"--owner-ref", "none",
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, armExpectedDir)
		})

		It("should not render RPS when disabled", func() {
			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", assetsInDir + "," + ppDirNoRPS,
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, ppNoRPSExpectedDir)
		})
	})

	Context("With pinned cluster resources", func() {
		It("Given default pinned infrastructure status, should render cpu partitioning configs", func() {

			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", defaultPinnedDir,
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, pinnedExpectedDir)
		})

		It("Given legacy SNO pinned infrastructure status, should render cpu partitioning configs", func() {

			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", snoLegacyPinnedDir,
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, pinnedExpectedDir)

		})
	})

	Context("With no MCPs manifest resources during bootstrap", func() {
		It("should render PerformanceProfile with default", func() {

			bootstrapPPDirs := []string{bootstrapPPDir, defaultPinnedDir}

			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(bootstrapPPDirs, ","),
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, path.Join(bootstrapExpectedDir, "no-mcp"))

		})
	})

	Context("With extra MCP manifest resources during bootstrap", func() {
		It("should render PerformanceProfile with default", func() {

			bootstrapPPDirs := []string{bootstrapPPDir, defaultPinnedDir, extraMCPDir}

			cmdline := []string{
				filepath.Join(binPath, "cluster-node-tuning-operator"),
				"render",
				"--asset-input-dir", strings.Join(bootstrapPPDirs, ","),
				"--asset-output-dir", assetsOutDir,
			}
			fmt.Fprintf(GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			runAndCompare(cmd, path.Join(bootstrapExpectedDir, "extra-mcp"))

		})
	})

	AfterEach(func() {
		cleanArtifacts()
	})

})

func createTempAssetsDir() string {
	assets, err := os.MkdirTemp("", "assets")
	Expect(err).ToNot(HaveOccurred())
	fmt.Printf("assets` output dir at: %q\n", assets)
	return assets
}

func cleanArtifacts() {
	os.RemoveAll(assetsOutDir)
}

func runAndCompare(cmd *exec.Cmd, dir string) {
	_, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred(), logStderr(err))

	outputAssetsFiles, err := os.ReadDir(assetsOutDir)
	Expect(err).ToNot(HaveOccurred())

	refPath := filepath.Join(testDataPath, "render-expected-output", dir)
	fmt.Fprintf(GinkgoWriter, "reference data at: %q\n", refPath)

	for _, f := range outputAssetsFiles {
		refData, err := os.ReadFile(filepath.Join(refPath, f.Name()))
		Expect(err).ToNot(HaveOccurred())

		data, err := os.ReadFile(filepath.Join(assetsOutDir, f.Name()))
		Expect(err).ToNot(HaveOccurred())

		diff, err := getFilesDiff(data, refData)
		Expect(err).ToNot(HaveOccurred())
		Expect(diff).To(BeZero(), "rendered %s file is not identical to its reference file; diff: %v",
			f.Name(),
			diff)
	}
}

func logStderr(err error) string {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Sprintf("error running the command: [[%s]]", exitErr.Stderr)
	}
	return fmt.Sprintf("error running the command: [[%s]]", err)
}
