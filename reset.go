package gestic

import (
	"fmt"
	"os"

	"github.com/ninjasphere/go-ninja/logger"
)

func ResetDevice() error {

	log := logger.GetLogger("Gestic Reset")

	log.Infof("resetting gestic device")

	err := writetofile("/sys/kernel/debug/omap_mux/mii1_rxdv", "2f")

	if err != nil {
		return fmt.Errorf("Unable to reset gestic device: %v", err)
	}

	err = writetofile("/sys/class/gpio/export", "100")

	if err != nil {
		return fmt.Errorf("Unable to write to export pin: %v", err)
	}

	err = writetofile("/sys/class/gpio/gpio100/direction", "out")

	if err != nil {
		return fmt.Errorf("Unable to reset gestic device: %v", err)
	}

	err = writetofile("/sys/class/gpio/gpio100/value", "0")

	if err != nil {
		return fmt.Errorf("Unable to reset gestic device: %v", err)
	}

	err = writetofile("/sys/class/gpio/gpio100/value", "1")

	if err != nil {
		return fmt.Errorf("Unable to reset gestic device: %v", err)
	}
	return nil
}

func writetofile(fn string, val string) error {

	df, err := os.OpenFile(fn, os.O_WRONLY|os.O_SYNC, 0666)

	if err != nil {
		return err
	}

	defer df.Close()

	if _, err = fmt.Fprintln(df, val); err != nil {
		return err
	}

	return nil
}
