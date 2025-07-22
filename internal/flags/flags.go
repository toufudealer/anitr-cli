package flags

import (
	"github.com/spf13/cobra"
)

type Flags struct {
	DisableRPC   bool
	CheckUpdate  bool
	PrintVersion bool
	RofiMode     bool
	RofiFlags    string
}

func NewFlagsCmd() (*cobra.Command, *Flags) {
	f := &Flags{}

	cmd := &cobra.Command{
		Use:   "anitr-cli",
		Short: "💫 Terminalden Türkçe anime izleme aracı",
	}

	cmd.PersistentFlags().BoolVar(&f.DisableRPC, "disable-rpc", false, "Discord Rich Presence özelliğini devre dışı bırakır.")
	cmd.PersistentFlags().BoolVar(&f.CheckUpdate, "update", false, "anitr-cli aracını en son sürüme günceller.")
	cmd.PersistentFlags().BoolVar(&f.PrintVersion, "version", false, "Versiyon bilgisi.")
	cmd.PersistentFlags().BoolVar(&f.RofiMode, "rofi", false, "Rofi arayüzü ile başlatır.")
	cmd.PersistentFlags().StringVar(&f.RofiFlags, "rofi-flags", "", "Rofi için flag'ler")

	return cmd, f
}
