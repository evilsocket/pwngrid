package mesh

import (
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
	"sort"
	"strconv"
	"time"
)

func ChannelHopping(iface string, chanList string, allChannels []int, hopPeriod int) {
	channels := []int{}
	for _, s := range str.Comma(chanList) {
		if ch, err := strconv.Atoi(s); err != nil {
			log.Fatal("%v", err)
		} else {
			channels = append(channels, ch)
		}
	}
	if len(channels) == 0 {
		channels = allChannels
	}
	sort.Ints(channels)

	go func() {
		period := time.Duration(hopPeriod) * time.Millisecond
		tick := time.NewTicker(period)

		log.Info("channel hopper started (period:%s channels:%v)", period, channels)

		loop := 0
		for _ = range tick.C {
			ch := channels[loop%len(channels)]
			// log.Debug("hopping on channel %d", ch)
			if err, out := SetChannel(iface, ch); err != nil {
				log.Error("%v: %s", err, out)
			}
			loop++
		}
	}()
}
