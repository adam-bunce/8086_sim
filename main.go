package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
)

func LoadInstructions(fileName string) (int, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	copy(MemoryValues.Bytes[0:len(file)], file)
	return len(file), nil
}

const (
	ShowInst int = iota
	ShowCycles
	ShowInstBytes
)

func main() {
	dumpMemory := flag.Bool("savemem", false, "save final memory state to .DATA file")
	dumpRegisters := flag.Bool("dumpreg", false, "output final register state")
	showInstructions := flag.Bool("print", false, "show instructions and their effect")
	showCycles := flag.Bool("cycles", false, "show # of cycles required to execute instruction")
	showInstBytes := flag.Bool("instbytes", false, "show the bytes that make up the instruction")
	flag.Parse()

	var programFileName string
	if len(flag.Args()) > 0 {
		programFileName = flag.Args()[0]
	} else {
		fmt.Println("Error: no file provided")
		flag.PrintDefaults()
		return
	}

	length, err := LoadInstructions(programFileName)
	if err != nil {
		fmt.Printf("failed to load instructions from %s\n%v\n", programFileName, err)
		return
	}

	At := 0
	instructions := Decode(MemoryValues.Bytes, &At)

	//  while the IP is within the range of memory keep doing stuff
	for ReadU16(RegisterValues[Register_ip], 0) < uint16(length) {
		Simulate(instructions[int(ReadU16(RegisterValues[Register_ip], 0))], []bool{*showInstructions, *showCycles, *showInstBytes})
	}

	if *dumpRegisters {
		fmt.Println()
		fmt.Println(RegisterValues)
		fmt.Println(CpuFlagValues)
	}

	if *dumpMemory {
		fileName := programFileName + "_memory.DATA"
		fmt.Println("saving program memory to", fileName)
		_ = os.Remove(fileName)
		err = os.WriteFile(fileName, MemoryValues.Bytes, 777)
		if err != nil {
			fmt.Println("Error: failed to save program's memory", err)
		}
	}
}
