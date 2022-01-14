package bed

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/warthog618/gpiod"
)

type WakeUp struct {
	fileName string
	chip     *gpiod.Chip

	dataLock sync.RWMutex
	data     *Data
}

func NewWakeup(chipName string, fileName string) (*WakeUp, error) {
	chip, err := gpiod.NewChip(chipName)

	if err != nil {
		return nil, fmt.Errorf("error initializing chip: %w", err)
	}

	var data Data

	// read existing data, ignore non-existing file
	storeData, err := os.ReadFile(fileName)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	if len(storeData) > 0 {
		if err = json.Unmarshal(storeData, &data); err != nil {
			return nil, fmt.Errorf("error unmarshalling file: %v", err)
		}
	}
	return &WakeUp{
		chip:     chip,
		data:     &data,
		fileName: fileName,
	}, nil
}

func (b *WakeUp) Run(ctx context.Context) error {

	line, err := b.chip.RequestLine(10, gpiod.WithEventHandler(b.HandlePush), gpiod.WithRisingEdge)

	if err != nil {
		return fmt.Errorf("error creating line watch: %v", err)
	}

	// wait for context
	<-ctx.Done()

	if err := line.Close(); err != nil {
		return fmt.Errorf("error closing line: %v", err)
	}
	return nil
}

func (b *WakeUp) Close() error {
	return b.chip.Close()
}

func (b *WakeUp) HandlePush(evt gpiod.LineEvent) {
	b.dataLock.Lock()
	defer b.dataLock.Unlock()

	b.data.Entries = append(b.data.Entries, &Entry{
		When: time.Now(),
	})

	marshalled, err := json.Marshal(b.data)
	if err != nil {
		log.Printf("error marshalling entry: %v", err)
		return
	}
	if err = ioutil.WriteFile(b.fileName, marshalled, os.ModePerm); err != nil {
		log.Printf("error writing file: %v", err)
	}
}

func (b *WakeUp) Entries() []*Entry {
	b.dataLock.RLock()
	defer b.dataLock.RUnlock()

	return append([]*Entry(nil), b.data.Entries...)
}
