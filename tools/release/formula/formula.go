package main

import (
	"fmt"
	"strings"
	"text/template"
)

var systems = []string{"darwin_arm64", "linux_arm64", "linux_amd64"}

type archive struct{ URL, Checksum string }

type formulaFields struct {
	Homepage                        string
	Version                         string
	DarwinARM, LinuxARM, LinuxIntel archive
}

var formulaTemplate = template.Must(template.New("formula").Parse(`class OnboardKit < Formula
  desc "Turns onboarding answers into a plan a newcomer walks through in the terminal"
  homepage "{{ .Homepage }}"
  version "{{ .Version }}"
  license "MIT"

  on_macos do
    on_arm do
      url "{{ .DarwinARM.URL }}"
      sha256 "{{ .DarwinARM.Checksum }}"
    end
  end

  on_linux do
    on_arm do
      url "{{ .LinuxARM.URL }}"
      sha256 "{{ .LinuxARM.Checksum }}"
    end
    on_intel do
      url "{{ .LinuxIntel.URL }}"
      sha256 "{{ .LinuxIntel.Checksum }}"
    end
  end

  def install
    bin.install "onboard-kit"
  end

  test do
    assert_match "onboard-kit", shell_output("#{bin}/onboard-kit --version")
  end
end
`))

func checksumsByFile(checksums string) map[string]string {
	found := map[string]string{}
	for _, line := range strings.Split(checksums, "\n") {
		if fields := strings.Fields(line); len(fields) == 2 {
			found[fields[1]] = fields[0]
		}
	}
	return found
}

func archiveName(tag, system string) string {
	return fmt.Sprintf("onboard-kit_%s_%s.tar.gz", tag, system)
}

func formula(tag, repository, checksums string) (string, error) {
	known := checksumsByFile(checksums)
	homepage := "https://github.com/" + repository
	found := map[string]archive{}
	for _, system := range systems {
		name := archiveName(tag, system)
		checksum, ok := known[name]
		if !ok {
			return "", fmt.Errorf("no checksum for %s", name)
		}
		found[system] = archive{URL: fmt.Sprintf("%s/releases/download/%s/%s", homepage, tag, name), Checksum: checksum}
	}
	fields := formulaFields{
		Homepage:   homepage,
		Version:    strings.TrimPrefix(tag, "v"),
		DarwinARM:  found["darwin_arm64"],
		LinuxARM:   found["linux_arm64"],
		LinuxIntel: found["linux_amd64"],
	}
	var text strings.Builder
	if err := formulaTemplate.Execute(&text, fields); err != nil {
		return "", err
	}
	return text.String(), nil
}
