package main

import (
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/mirror/chunk"

	"github.com/pterm/pterm"
)

type States struct {
	states string
}

func main() {
	myMap := map[*chunk.LegacyBlock]*States{}

	olderBlock := chunk.LegacyBlocks
	for key, value := range olderBlock {
		if value == nil {
			pterm.Info.Printf("%v is nil\n", key)
			continue
		}
		myMap[value] = &States{}
	}

	for key, value := range myMap {
		runtimeID, success := chunk.LegacyBlockToRuntimeID(
			key.Name,
			key.Val,
		)
		if !success {
			pterm.Error.Printf("ERR 1 | %#v failed\n", key)
			continue
		}

		correctName, states, success := chunk.RuntimeIDToState(runtimeID)
		if !success {
			pterm.Error.Printf(
				"ERR 2 | %v(runtimeId %v) failed\n",
				key,
				runtimeID,
			)
			continue
		}

		got, err := mcstructure.ConvertCompoundToString(states, true)
		if err != nil {
			pterm.Error.Printf(
				"ERR 3 | %v\n",
				err,
			)
			continue
		}

		key.Name = correctName
		*value = States{states: got}
	}

	strPool := []string{}
	for key, value := range myMap {
		nameWithData, err := json.Marshal(fmt.Sprintf("%v|%d", key.Name, key.Val))
		if err != nil {
			pterm.Error.Printf(
				"ERR 4 | %v\n",
				err,
			)
			continue
		}

		blockStates, err := json.Marshal(value.states)
		if err != nil {
			pterm.Error.Printf(
				"ERR 4 | %v\n",
				err,
			)
			continue
		}

		strPool = append(
			strPool,
			fmt.Sprintf(`%v: %v`, string(nameWithData), string(blockStates)),
		)
	}

	str := `{`
	for _, value := range strPool {
		str = str + value + `, `
	}
	str = str[:len(str)-2]
	str = str + `}`

	file, _ := os.OpenFile("ans.json", os.O_CREATE|os.O_WRONLY, 2)
	file.Write([]byte(str))
	file.Close()
}
