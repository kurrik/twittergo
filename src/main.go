// Copyright 2011 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package main

package main

import (
	"twittergo"
	"fmt"
	"reflect"
	"os"
	"path"
	"path/filepath"
	"gob"
	"bufio"
	"strings"
	"exec"
	"bytes"
)

// Pretty-prints a value to the terminal.
func PrintValue(value reflect.Value, indent string) {
	value = reflect.Indirect(value)
	if value.IsValid() {
		switch value.Kind() {
		case reflect.Array, reflect.Slice:
			PrintArray(value.Interface(), indent)
		case reflect.Struct:
			PrintStruct(value.Interface(), indent)
		case reflect.Bool:
			fmt.Println(indent, value.Bool())
		case reflect.Uint32, reflect.Uint64:
			fmt.Println(indent, value.Uint())
		case reflect.Int, reflect.Int32, reflect.Int64:
			fmt.Println(indent, value.Int())
		default:
			fmt.Println(indent, value.String(), value.Kind())
		}
	}
}

// Pretty-prints an array to the terminal.
func PrintArray(item interface{}, indent string) {
	itemValue := reflect.ValueOf(item)
	for i := 0; i < itemValue.Len(); i++ {
		value := itemValue.Index(i)
		fmt.Println(indent, "[")
		PrintValue(value, indent+"    ")
		fmt.Println(indent, "]")
	}
}

// Pretty-prints a struct to the terminal.
func PrintStruct(item interface{}, indent string) {
	itemType := reflect.TypeOf(item)
	itemValue := reflect.ValueOf(item)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		value := itemValue.FieldByName(field.Name)
		value = reflect.Indirect(value)
		if value.IsValid() {
			fmt.Println(indent, field.Name, field.Type)
			PrintValue(value, indent+"    ")
		}
	}
}

// Prompts the user and returns their input as a string.
func PromptUser(prompt string) string {
	if prompt == "" {
		fmt.Print("> ")
	} else {
		fmt.Print(prompt + ": ")
	}
	reader := bufio.NewReader(os.Stdin)
	buffer := new(bytes.Buffer)
	for {
		char, err := reader.ReadByte()
		if err != nil || char == '\n' {
			break
		}
		buffer.WriteByte(char)
	}
	fmt.Println("GOT", buffer.String())
	return buffer.String()
}

// Encodes a /bin/stty state as a string.
type InputState string

// Sets the terminal to raw mode.
func SetRaw() InputState {
	command := exec.Command("/bin/stty", "-g")
	command.Stdin = os.Stdin
	currentState, err := command.Output()
	if err != nil {
		return InputState("")
	}
	command = exec.Command("/bin/stty", "raw")
	command.Stdin = os.Stdin
	err = command.Run()
	if err != nil {
		return InputState("")
	}
	return InputState(currentState[0 : len(currentState)-1])
}

// Sets the terminal back into cooked mode.
func (state InputState) Revert() {
	var command *exec.Cmd
	if state == InputState("") {
		// No state change happened
		return
	} else if state == InputState("cooked") {
		// Still playing around with this
		// Somehow istrip got enabled and killed all UTF-8 pasting
		// on my system.  Not sure what I want to do here.
		command = exec.Command("/bin/stty", "cooked", "-istrip")
	} else {
		command = exec.Command("/bin/stty", string(state))
	}
	command.Stdin = os.Stdin
	err := command.Run()
	if err != nil {
		fmt.Println("Error resetting terminal state")
	}
}

// Key codes of interest.
const (
	COMBO_CONTROL_C = 3
	COMBO_CONTROL_D = 4
	COMBO_CONTROL_E = 5
	ESCAPE          = 27
	PAGE_CONTROLS   = 91
	KEY_ENTER       = 13
	KEY_BACKSPACE   = 127
	KEY_UP          = 65
	KEY_DOWN        = 66
	KEY_LEFT        = 68
	KEY_RIGHT       = 67
)

// Obtains the name of a command in a very convoluted way.
// Goes into raw terminal mode to support arrowing through commands.
func GetCommand(registry *CommandRegistry) string {
	defer SetRaw().Revert()
	reader := bufio.NewReader(os.Stdin)
	buffer := new(bytes.Buffer)
	unknown := new(bytes.Buffer)
	fmt.Print("> ")
	registryIndex := -1
	command := ""
	for command == "" {
		char, _ := reader.ReadByte()
		switch char {
		case KEY_ENTER:
			command = buffer.String()
		case KEY_BACKSPACE:
			if buffer.Len() > 0 {
				buffer.Truncate(buffer.Len() - 1)
			}
		case COMBO_CONTROL_C, COMBO_CONTROL_D, COMBO_CONTROL_E:
			command = "exit"
		case ESCAPE:
			char, _ := reader.ReadByte()
			switch char {
			case PAGE_CONTROLS:
				char, _ := reader.ReadByte()
				switch char {
				case KEY_LEFT:
					fmt.Print("\033[D")
				case KEY_RIGHT:
					fmt.Print("\033[C")
				case KEY_DOWN:
					registryIndex = (registryIndex + 1) % len(registry.Names)
					buffer.Reset()
					buffer.WriteString(registry.Names[registryIndex])
				case KEY_UP:
					if registryIndex >= 0 {
						registryIndex -= 1
					}
					registryIndex = (registryIndex + len(registry.Names)) % len(registry.Names)
					buffer.Reset()
					buffer.WriteString(registry.Names[registryIndex])
				default:
					unknown.WriteByte(uint8(char))
				}
			default:
				unknown.WriteByte(uint8(char))
			}
		default:
			buffer.WriteByte(uint8(char))
			unknown.WriteByte(uint8(char))
		}
		fmt.Print("\033[2K\033[0G")
		//fmt.Printf("[%v][%v] > %v", char, unknown, buffer.String())
		fmt.Printf("> %v", buffer.String())
	}
	fmt.Printf("\033[0G\n")
	return command
}

// Returns the directory path of the directory this executable is stored in.
func GetExecutableDirectory() string {
	wd, _ := os.Getwd()
	dir, _ := path.Split(path.Clean(path.Join(wd, os.Args[0])))
	return dir
}

// Saves a serialized version of an interface to the specified path.
// Sets the file mask to 600.
func SaveConfig(path string, config interface{}) os.Error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	encoder.Encode(config)
	return os.Chmod(path, 384) // Same as chmod 600
}

// Loads a serialized version of an interface from the specified path.
func LoadConfig(path string, out interface{}) os.Error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil
}

// Obtains an OAuthConfig.
// Loads an existing client configuration, or prompts the user if none is found.
func GetClientConfig() (*twittergo.OAuthConfig, os.Error) {
	configPath := path.Join(GetExecutableDirectory(), "client.config")
	var config twittergo.OAuthConfig
	if err := LoadConfig(configPath, &config); err != nil {
		fmt.Println("Could not find client configuration at", configPath)
		key := PromptUser("Enter your Twitter Consumer Key")
		secret := PromptUser("Enter your Twitter Consumer Secret")
		config := twittergo.NewOAuthConfig(key, secret, "oob")
		err = SaveConfig(configPath, config)
		if err != nil {
			fmt.Println("Could not save configuration to", configPath)
			return nil, err
		}
	}
	return &config, nil
}

// Obtains an OAuthUserConfig.
// Prompts the user to load a specified config, or starts a new OAuth flow.
func GetUserConfig(config *twittergo.OAuthConfig) (*twittergo.OAuthUserConfig, os.Error) {
	dirPath := GetExecutableDirectory()
	globPath := dirPath + "*.twitter"
	matches, _ := filepath.Glob(globPath)
	names := make([]string, len(matches))
	for i, match := range matches {
		_, names[i] = filepath.Split(match)
		names[i] = strings.Replace(names[i], ".twitter", "", 1)
	}
	var choice int = 0
	if len(names) > 0 {
		for {
			fmt.Println("Choose an account to use:")
			fmt.Println("  0. Authorize new account")
			for i, name := range names {
				fmt.Printf("  %v. %v\n", i+1, name)
			}
			input := PromptUser("Choice")
			_, err := fmt.Sscanf(input, "%d", &choice)
			if err == nil && choice >= 0 && choice <= len(names) {
				break
			}
			fmt.Println("There was a problem parsing your input.")
		}
	}
	var userConfig twittergo.OAuthUserConfig
	if choice > 0 {
		configPath := path.Join(dirPath, names[choice-1]+".twitter")
		if err := LoadConfig(configPath, &userConfig); err == nil {
			if userConfig.AccessTokenKey != "" {
				return &userConfig, nil
			} else {
				fmt.Println("Config was not initialized")
			}
		} else {
			fmt.Println("Problem loading the requested config.")
		}
	} else {
		userConfig = twittergo.OAuthUserConfig{}
	}
	fmt.Println("Starting a new auth flow")
	client := twittergo.NewClient(config, &userConfig)
	if err := client.GetRequestToken(); err != nil {
		return nil, err
	}
	url, err := client.GetAuthorizeUrl()
	if err != nil {
		return nil, err
	}
	fmt.Println("Please visit this URL in your browser:", url)
	pin := PromptUser("Please input the PIN displayed")
	if err := client.GetAccessToken(userConfig.RequestTokenKey, pin); err != nil {
		return nil, err
	}
	configPath := path.Join(dirPath, userConfig.AccessValues["screen_name"][0]+".twitter")
	err = SaveConfig(configPath, userConfig)
	if err != nil {
		return nil, err
	}
	return &userConfig, nil
}

// Represents a command which may be input in the terminal.
type Command struct {
	Name     string
	Help     string
	Function func()
}

// Represents a directory of all available commands.
type CommandRegistry struct {
	Commands map[string]Command
	Names    []string
}

// Register a new command to be used.
func (c *CommandRegistry) Register(command Command) {
	if c.Commands == nil {
		c.Commands = map[string]Command{}
	}
	c.Commands[command.Name] = command
	c.Names = append(c.Names, command.Name)
}

// Print a help page detailing all of the commands.
func (c *CommandRegistry) PrintHelp() {
	for _, name := range c.Names {
		command := c.Commands[name]
		fmt.Printf("\t%-30v   %v\n", name, command.Help)
	}
}

// Execute the command with the given name.
func (c *CommandRegistry) Execute(name string) {
	command, present := c.Commands[name]
	if present {
		command.Function()
	} else {
		fmt.Println("Invalid command:", name)
	}
}

func main() {
	// Fixes a broken console if the executable was killed.
	//InputState("cooked").Revert()

	var (
		exit         bool = false
		err          os.Error
		arrayOutput  interface{}
		structOutput interface{}
		client       *twittergo.Client
	)

	clientConfig, err := GetClientConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	userConfig, err := GetUserConfig(clientConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client = twittergo.NewClient(clientConfig, userConfig)
	registry := CommandRegistry{}

	registry.Register(Command{"help", "Prints this message", func() {
		registry.PrintHelp()
	}})

	registry.Register(Command{"exit", "Exits", func() {
		exit = true
	}})

	registry.Register(Command{"status/public_timeline", "", func() {
		arrayOutput, err = client.GetPublicTimeline(nil)
	}})

	registry.Register(Command{"status/retweeted_by_me", "", func() {
		arrayOutput, err = client.GetRetweetedByMe(nil)
	}})

	registry.Register(Command{"statuses/:id/retweeted_by", "", func() {
		id := PromptUser("Tweet ID")
		arrayOutput, err = client.GetStatusRetweetedBy(id, true, nil)
	}})

	registry.Register(Command{"statuses/update", "Posts a tweet", func() {
		status := PromptUser("What's happening?")
		structOutput, err = client.Update(status, nil)
	}})

	var command string
	for exit == false {
		fmt.Println("\nEnter a command (\"help\" for options, up/down to cycle through commands)")
		command = GetCommand(&registry)
		registry.Execute(command)

		if err != nil {
			fmt.Println("Error:", err)
			err = nil
		}
		if arrayOutput != nil {
			PrintArray(arrayOutput, "")
			arrayOutput = nil
		}
		if structOutput != nil {
			PrintStruct(structOutput, "")
			structOutput = nil
		}
	}
}
