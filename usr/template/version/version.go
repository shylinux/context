package cli

var version = struct {
	init []string
	time string
	host string
	self int
}{
	[]string{"2017-11-01 01:02:03", "2019-07-13 18:02:21"},
	`{{options . "time"}}`, `{{options . "host"}}`, {{options . "self"}},
}
