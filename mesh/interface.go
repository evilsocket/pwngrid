package mesh

import (
	"bufio"
	"fmt"
	"github.com/evilsocket/pwngrid/utils"
	"regexp"
	"strconv"
	"strings"
)

var chanParser = regexp.MustCompile(`^\s+Channel.([0-9]+)\s+:\s+([0-9\.]+)\s+GHz.*$`)

func ActivateInterface(name string) error {
	if out, err := utils.Exec("ifconfig", []string{name, "up"}); err != nil {
		return err
	} else if out != "" {
		return fmt.Errorf("unexpected output while activating interface %s: %s", name, out)
	}
	return nil
}

func SetChannel(iface string, channel int) (error, string) {
	if out, err := utils.Exec("iwconfig", []string{iface, "channel", fmt.Sprintf("%d", channel)}); err != nil {
		return err, out
	} else if out != "" {
		return fmt.Errorf("unexpected output while setting interface %s to channel %d: %s", iface, channel, out), out
	} else {
		return nil, out
	}
}

func SupportedChannels(iface string) ([]int, error) {
	out, err := utils.Exec("iwlist", []string{iface, "freq"})
	if err != nil {
		return nil, err
	}

	channels := []int{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := chanParser.FindStringSubmatch(line); len(matches) == 3 {
			if channel, err := strconv.ParseInt(matches[1], 10, 32); err == nil {
				channels = append(channels, int(channel))
			}
		}
	}

	return channels, nil
}
