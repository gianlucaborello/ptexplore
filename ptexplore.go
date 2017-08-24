package ptexplore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	humanize "github.com/dustin/go-humanize"
)

const pageMapReadChunk uint64 = 8

const pfnMask uint64 = 0x7FFFFFFFFFFFFF

const (
	pageSoftDirty uint64 = 1 << 55
	pageExclusive uint64 = 1 << 56
	pageFile      uint64 = 1 << 61
	pageSwapped   uint64 = 1 << 62
	pagePresent   uint64 = 1 << 63
)

var pageFlagsMap = map[int64]string{
	1 << 0:  "LOCKED",
	1 << 1:  "ERROR",
	1 << 2:  "REFERENCED",
	1 << 3:  "UPTODATE",
	1 << 4:  "DIRTY",
	1 << 5:  "LRU",
	1 << 6:  "ACTIVE",
	1 << 7:  "SLAB",
	1 << 8:  "WRITEBACK",
	1 << 9:  "RECLAIM",
	1 << 10: "BUDDY",
	1 << 11: "MMAP",
	1 << 12: "ANON",
	1 << 13: "SWAPCACHE",
	1 << 14: "SWAPBACKED",
	1 << 15: "COMPOUND_HEAD",
	1 << 16: "COMPOUND_TAIL",
	1 << 17: "HUGE",
	1 << 18: "UNEVICTABLE",
	1 << 19: "HWPOISON",
	1 << 20: "NOPAGE",
	1 << 21: "KSM",
	1 << 22: "THP",
	1 << 23: "BALLOON",
	1 << 24: "ZERO_PAGE",
	1 << 25: "IDLE",
}

var pageSize = uint64(os.Getpagesize())

type memArea struct {
	start    uint64
	end      uint64
	pathName string
}

func (area *memArea) String() string {
	return fmt.Sprintf("Area '%s', range %0#16x - %0#16x (%v)",
		area.pathName, area.start, area.end, humanize.Bytes(area.end-area.start))
}

type PtExplorerState struct {
	pageMapFile   *os.File
	pageCountFile *os.File
	pageFlagsFile *os.File
	memAreas      []memArea
}

func (p *PtExplorerState) OpenSystemFiles(pid int) error {
	pageMap := fmt.Sprintf("/proc/%v/pagemap", pid)
	var err error
	p.pageMapFile, err = os.Open(pageMap)
	if err != nil {
		return err
	}

	p.pageCountFile, err = os.Open("/proc/kpagecount")
	if err != nil {
		return err
	}

	p.pageFlagsFile, err = os.Open("/proc/kpageflags")
	if err != nil {
		return err
	}

	return nil
}

func (p *PtExplorerState) ParseMemAreas(pid int) error {
	pmap := fmt.Sprintf("/proc/%v/maps", pid)
	content, err := ioutil.ReadFile(pmap)
	if err != nil {
		return err
	}

	p.memAreas = make([]memArea, 0)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)

		var area memArea
		if len(fields) == 0 {
			continue
		}

		memRange := strings.Split(fields[0], "-")
		area.start, err = strconv.ParseUint(memRange[0], 16, 64)
		if err != nil {
			return err
		}

		area.end, err = strconv.ParseUint(memRange[1], 16, 64)
		if err != nil {
			return err
		}

		if len(fields) == 6 {
			area.pathName = fields[5]
		} else {
			area.pathName = "anonymous"
		}

		p.memAreas = append(p.memAreas, area)
	}

	return nil
}

func (p *PtExplorerState) PrintAreas(areaFilters string, addressFilter uint64, quiet bool) error {
	areas := p.getAreasToPrint(areaFilters)

	for _, area := range areas {
		err := p.printArea(area, addressFilter, quiet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PtExplorerState) getAreasToPrint(filters string) []memArea {
	var areasToParse []memArea
	if filters != "" {
		filterList := strings.Split(filters, ",")
		for _, filter := range filterList {
			for _, area := range p.memAreas {
				if strings.Contains(area.pathName, filter) {
					areasToParse = append(areasToParse, area)
				}
			}
		}
	} else {
		areasToParse = p.memAreas
	}

	return areasToParse
}

func (p *PtExplorerState) printArea(area memArea, addressFilter uint64, quiet bool) error {
	if addressFilter != 0 && (addressFilter < area.start || addressFilter >= area.end) {
		return nil
	}

	if !quiet {
		fmt.Printf("\n")
	}

	fmt.Printf("%v\n", area.String())

	if !quiet {
		fmt.Printf("\n")
	}

	if quiet {
		return nil
	}

	nonMapped := 0

	for addr := area.start; addr < area.end; addr += pageSize {
		if addressFilter != 0 && (addressFilter < addr || addressFilter >= addr+pageSize) {
			continue
		}

		err := p.printPage(addr, &nonMapped)
		if err != nil {
			return err
		}
	}

	printNonMapped(nonMapped)

	return nil
}

func (p *PtExplorerState) printPage(address uint64, nonMapped *int) error {

	pageEntry, err := p.getPageEntry(address)
	if err != nil {
		return err
	}

	if pageEntry&pagePresent != 0 || pageEntry&pageSwapped != 0 {
		printNonMapped(*nonMapped)
		*nonMapped = 0
	}

	if pageEntry&pagePresent != 0 {
		fmt.Printf("%0#16x: ", address)

		pfn := pageEntry & pfnMask
		fmt.Printf("physical address: %0#16x ", pfn*pageSize)

		if pageEntry&pageFile != 0 {
			fmt.Printf("file-page ")
		}

		if pageEntry&pageExclusive != 0 {
			fmt.Printf("exclusive ")
		}

		if pageEntry&pageSoftDirty != 0 {
			fmt.Printf("soft-dirty ")
		}

		pageCount, err := p.getPageCount(pfn)
		if err != nil {
			return err
		}

		fmt.Printf("count:%v ", pageCount)

		pageFlags, err := p.getPageFlags(pfn)
		if err != nil {
			return err
		}

		fmt.Printf("flags:")
		first := true
		var i uint
		for i = 0; i < 64; i++ {
			if pageFlags&(1<<i) != 0 {
				flag, found := pageFlagsMap[1<<i]
				if !found {
					continue
				}
				if !first {
					fmt.Printf(",")
				} else {
					first = false
				}

				fmt.Printf(flag)
			}
		}
		fmt.Printf("\n")
	} else if pageEntry&pageSwapped != 0 {
		fmt.Printf("%0#16x: ", address)
		fmt.Printf("swapped ")
		fmt.Printf("\n")
	} else {
		*nonMapped++
	}

	return nil
}

func (p *PtExplorerState) getPageEntry(address uint64) (entry uint64, err error) {
	pageNumber := address / pageSize
	off := int64(pageNumber * pageMapReadChunk)
	buf := make([]byte, pageMapReadChunk)
	_, err = p.pageMapFile.ReadAt(buf, off)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &entry)
	if err != nil {
		return 0, err
	}

	return entry, nil
}

func (p *PtExplorerState) getPageFlags(pfn uint64) (count uint64, err error) {
	buf := make([]byte, pageMapReadChunk)
	off := int64(pfn * pageMapReadChunk)
	_, err = p.pageFlagsFile.ReadAt(buf, off)
	if err != nil {
		return 0, err
	}
	var flags uint64
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &flags)
	if err != nil {
		return 0, err
	}

	return flags, nil
}

func (p *PtExplorerState) getPageCount(pfn uint64) (count uint64, err error) {
	buf := make([]byte, pageMapReadChunk)
	off := int64(pfn * pageMapReadChunk)

	_, err = p.pageCountFile.ReadAt(buf, off)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func printNonMapped(count int) {
	if count != 0 {
		fmt.Printf("... %v non mapped pages ...\n", count)
	}
}
