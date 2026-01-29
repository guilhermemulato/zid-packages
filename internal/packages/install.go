package packages

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func installBundle(pkg Package) error {
	if pkg.BundleURL == "" || pkg.InstallScriptGlob == "" {
		return errors.New("bundle url ou install script nao definido")
	}
	downloader, err := pickDownloader()
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp("/tmp", "zid-packages-install.")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	bundle := filepath.Join(tmpDir, "bundle.tar.gz")
	if err := downloadFile(downloader, pkg.BundleURL, bundle); err != nil {
		return err
	}
	extractDir := filepath.Join(tmpDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return err
	}

	if err := exec.Command("/usr/bin/tar", "-xzf", bundle, "-C", extractDir).Run(); err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(extractDir, pkg.InstallScriptGlob))
	if err != nil || len(matches) == 0 {
		return errors.New("install.sh nao encontrado no bundle")
	}
	installScript := matches[0]
	if err := os.Chmod(installScript, 0755); err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runUpdate(cmdPath string) error {
	cmd := exec.Command("/bin/sh", cmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pickDownloader() (string, error) {
	if _, err := exec.LookPath("fetch"); err == nil {
		return "fetch", nil
	}
	if _, err := exec.LookPath("curl"); err == nil {
		return "curl", nil
	}
	return "", errors.New("fetch/curl nao encontrado")
}

func downloadFile(downloader, url, dest string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("url vazio")
	}
	if downloader == "fetch" {
		cmd := exec.Command("fetch", "-o", dest, url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("curl", "-fL", "-o", dest, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
