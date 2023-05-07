package cmd

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const banner = `XXXXXXX       XXXXXXIIIIIIIIII
X:::::X       X:::::I::::::::I
X:::::X       X:::::I::::::::I
X::::::X     X::::::II::::::II
XXX:::::X   X:::::XXX I::::I    mmmmmmm    mmmmmmm    aaaaaaaaaaaaa    ggggggggg   ggggg   eeeeeeeeeeee   rrrrr   rrrrrrrrr
   X:::::X X:::::X    I::::I  mm:::::::m  m:::::::mm  a::::::::::::a  g:::::::::ggg::::g ee::::::::::::ee r::::rrr:::::::::r
    X:::::X:::::X     I::::I m::::::::::mm::::::::::m aaaaaaaaa:::::ag:::::::::::::::::ge::::::eeeee:::::er:::::::::::::::::r
     X:::::::::X      I::::I m::::::::::::::::::::::m          a::::g::::::ggggg::::::ge::::::e     e:::::rr::::::rrrrr::::::r
     X:::::::::X      I::::I m:::::mmm::::::mmm:::::m   aaaaaaa:::::g:::::g     g:::::ge:::::::eeeee::::::er:::::r     r:::::r
    X:::::X:::::X     I::::I m::::m   m::::m   m::::m aa::::::::::::g:::::g     g:::::ge:::::::::::::::::e r:::::r     rrrrrrr
   X:::::X X:::::X    I::::I m::::m   m::::m   m::::ma::::aaaa::::::g:::::g     g:::::ge::::::eeeeeeeeeee  r:::::r
XXX:::::X   X:::::XXX I::::I m::::m   m::::m   m::::a::::a    a:::::g::::::g    g:::::ge:::::::e           r:::::r
X::::::X     X::::::II::::::Im::::m   m::::m   m::::a::::a    a:::::g:::::::ggggg:::::ge::::::::e          r:::::r
X:::::X       X:::::I::::::::m::::m   m::::m   m::::a:::::aaaa::::::ag::::::::::::::::g e::::::::eeeeeeee  r:::::r
X:::::X       X:::::I::::::::m::::m   m::::m   m::::ma::::::::::aa:::agg::::::::::::::g  ee:::::::::::::e  r:::::r
XXXXXXX       XXXXXXIIIIIIIIImmmmmm   mmmmmm   mmmmmm aaaaaaaaaa  aaaa  gggggggg::::::g    eeeeeeeeeeeeee  rrrrrrr
                                                                                g:::::g
                                                                    gggggg      g:::::g
                                                                    g:::::gg   gg:::::g
                                                                     g::::::ggg:::::::g
                                                                      gg:::::::::::::g
                                                                        ggg::::::ggg
                                                                           gggggg`

var (
	version   = ""
	gitHash   = ""
	buildDate = ""
)

// versionCmd represents the worker command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version of XImager",
	Run: func(_ *cobra.Command, _ []string) {
		color.Cyan(banner)
		fmt.Printf("Version:     %s\n", version)
		fmt.Printf("GoVersion:   %s\n", runtime.Version())
		fmt.Printf("Platform:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("BuildDate:   %s\n", buildDate)
		fmt.Printf("GitCommit:   %s\n", gitHash)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
