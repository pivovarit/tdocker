package ui

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

func checkUpdateCmd(current string) tea.Cmd {
	return func() tea.Msg {
		if current == "dev" || current == "" {
			return nil
		}
		time.Sleep(2 * time.Second)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/pivovarit/tdocker/releases/latest")
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil
		}
		if isNewerVersion(release.TagName, current) {
			return updateAvailableMsg{version: release.TagName}
		}
		return nil
	}
}

func isNewerVersion(latest, current string) bool {
	lv := parseSemver(latest)
	cv := parseSemver(current)
	if lv == nil || cv == nil {
		return false
	}
	for i := range lv {
		if lv[i] != cv[i] {
			return lv[i] > cv[i]
		}
	}
	return false
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}
