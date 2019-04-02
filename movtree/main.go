package main

import (
	"fmt"
	"io"
	"os"

	"github.com/theaaf/qtff-go"
)

type AtomTypeInfo struct {
	IsContainer bool
	Print       func(data interface{}, indent string)
}

var atomTypes = map[string]AtomTypeInfo{
	"moov": AtomTypeInfo{IsContainer: true},
	"trak": AtomTypeInfo{IsContainer: true},
	"mdia": AtomTypeInfo{IsContainer: true},
	"minf": AtomTypeInfo{IsContainer: true},
	"stbl": AtomTypeInfo{IsContainer: true},
	"dinf": AtomTypeInfo{IsContainer: true},
	"tref": AtomTypeInfo{IsContainer: true},
	"mdhd": AtomTypeInfo{
		Print: func(iface interface{}, indent string) {
			data := iface.(*qtff.MediaHeaderData)
			fmt.Printf(indent+"time scale: %v\n", data.TimeScale)
			fmt.Printf(indent+"duration: %v\n", data.Duration)
		},
	},
	"hdlr": AtomTypeInfo{
		Print: func(iface interface{}, indent string) {
			data := iface.(*qtff.HandlerReferenceData)
			fmt.Printf(indent+"component type: %v\n", data.ComponentType)
			fmt.Printf(indent+"component subtype: %v\n", data.ComponentSubtype)
		},
	},
	"stco": AtomTypeInfo{
		Print: func(iface interface{}, indent string) {
			data := iface.(*qtff.ChunkOffsetData)
			fmt.Printf(indent+"number of entries: %v\n", data.NumberOfEntries)
			if len(data.Offsets) > 0 {
				fmt.Printf(indent+"first chunk offset: %v\n", data.Offsets[0])
			}
		},
	},
	"co64": AtomTypeInfo{
		Print: func(iface interface{}, indent string) {
			data := iface.(*qtff.ChunkOffset64Data)
			fmt.Printf(indent+"number of entries: %v\n", data.NumberOfEntries)
			if len(data.Offsets) > 0 {
				fmt.Printf(indent+"first chunk offset: %v\n", data.Offsets[0])
			}
		},
	},
	"stsz": AtomTypeInfo{
		Print: func(iface interface{}, indent string) {
			data := iface.(*qtff.SampleSizeData)
			if data.ConstantSampleSize != 0 {
				fmt.Printf(indent+"constant sample size: %v\n", data.ConstantSampleSize)
			} else {
				fmt.Printf(indent+"number of entries: %v\n", data.NumberOfEntries)
				if len(data.SampleSizes) > 0 {
					fmt.Printf(indent+"first sample size: %v\n", data.SampleSizes[0])
				}
			}
		},
	},
}

func tree(r io.ReaderAt, indent string) error {
	ar := qtff.NewAtomReader(r)

	for atom := ar.Next(); atom != nil; atom = ar.Next() {
		fmt.Printf(indent+"%s (%v bytes)\n", atom.Type.String(), atom.Size)
		if info, ok := atomTypes[atom.Type.String()]; !ok {
			continue
		} else if info.IsContainer {
			if err := tree(atom.Data, indent+"  "); err != nil {
				return err
			}
		} else if info.Print != nil {
			data, err := atom.ParseData()
			if err != nil {
				return err
			}
			info.Print(data, indent+"  ")
		}
	}

	return ar.Error()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("expected input argument")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	if err := tree(f, ""); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
