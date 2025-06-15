// Author: Ervins Strauhmanis
// License: MIT

package gui

import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// ErrorReport handles error reporting and debugging information
type ErrorReport struct {
	logger    *logrus.Logger
	debugMode bool
}

// NewErrorReport creates a new error report component
func NewErrorReport(logger *logrus.Logger, debugMode bool) *ErrorReport {
	return &ErrorReport{
		logger:    logger,
		debugMode: debugMode,
	}
}

// ShowError displays an error dialog with debugging information
func (er *ErrorReport) ShowError(err error) {
	if err == nil {
		return
	}

	er.logger.WithError(err).Error("Error occurred")

	if er.debugMode {
		er.showDetailedErrorDialog(err)
	} else {
		er.showSimpleErrorDialog(err)
	}
}

// showSimpleErrorDialog shows a simple error dialog for end users
func (er *ErrorReport) showSimpleErrorDialog(err error) {
	dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
}

// showDetailedErrorDialog shows a detailed error dialog with debugging information
func (er *ErrorReport) showDetailedErrorDialog(err error) {
	// Get stack trace
	buf := make([]byte, 4096)
	stackSize := runtime.Stack(buf, false)
	stackTrace := string(buf[:stackSize])

	// Create error details
	errorDetails := fmt.Sprintf(`Error: %s

Time: %s
Go Version: %s
OS: %s
Architecture: %s

Stack Trace:
%s`, 
		err.Error(),
		time.Now().Format("2006-01-02 15:04:05"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		stackTrace,
	)

	// Create UI components
	errorLabel := widget.NewLabel("An error occurred:")
	errorText := widget.NewLabel(err.Error())
	errorText.Wrapping = fyne.TextWrapWord

	detailsLabel := widget.NewLabel("Technical Details:")
	detailsEntry := widget.NewMultiLineEntry()
	detailsEntry.SetText(errorDetails)
	detailsEntry.Resize(fyne.NewSize(600, 300))

	// Create buttons
	copyButton := widget.NewButton("Copy Details", func() {
		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(errorDetails)
	})

	reportButton := widget.NewButton("Report Issue", func() {
		er.showReportDialog(err, errorDetails)
	})

	closeButton := widget.NewButton("Close", nil)

	// Create layout
	content := container.NewVBox(
		errorLabel,
		errorText,
		widget.NewSeparator(),
		detailsLabel,
		container.NewScroll(detailsEntry),
		container.NewHBox(
			copyButton,
			reportButton,
			closeButton,
		),
	)

	// Create dialog
	errorDialog := dialog.NewCustom("Error Details", "", content, fyne.CurrentApp().Driver().AllWindows()[0])
	errorDialog.Resize(fyne.NewSize(700, 500))

	closeButton.OnTapped = func() {
		errorDialog.Hide()
	}

	errorDialog.Show()
}

// showReportDialog shows a dialog for reporting issues
func (er *ErrorReport) showReportDialog(err error, technicalDetails string) {
	// Create form fields
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Your name (optional)")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Your email (optional)")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Describe what you were doing when the error occurred...")
	descriptionEntry.Resize(fyne.NewSize(400, 150))

	stepsEntry := widget.NewMultiLineEntry()
	stepsEntry.SetPlaceHolder("Steps to reproduce the error...")
	stepsEntry.Resize(fyne.NewSize(400, 100))

	includeDetailsCheck := widget.NewCheck("Include technical details", nil)
	includeDetailsCheck.SetChecked(true)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name:", Widget: nameEntry},
			{Text: "Email:", Widget: emailEntry},
			{Text: "Description:", Widget: descriptionEntry},
			{Text: "Steps to reproduce:", Widget: stepsEntry},
			{Text: "", Widget: includeDetailsCheck},
		},
	}

	// Create dialog
	reportDialog := dialog.NewCustomConfirm("Report Issue", "Submit", "Cancel",
		form,
		func(submit bool) {
			if !submit {
				return
			}

			// Prepare report content
			report := fmt.Sprintf(`Bug Report

Name: %s
Email: %s

Description:
%s

Steps to Reproduce:
%s

Error: %s
`, 
				nameEntry.Text,
				emailEntry.Text,
				descriptionEntry.Text,
				stepsEntry.Text,
				err.Error(),
			)

			if includeDetailsCheck.Checked {
				report += fmt.Sprintf("\nTechnical Details:\n%s", technicalDetails)
			}

			// In a real application, you would send this to your bug tracking system
			// For now, we'll just copy it to clipboard and show instructions
			clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
			clipboard.SetContent(report)

			dialog.ShowInformation("Report Prepared", 
				"Bug report has been copied to your clipboard.\n\nPlease paste it into an email and send to:\nsupport@example.com", 
				fyne.CurrentApp().Driver().AllWindows()[0])
			
			er.logger.WithFields(logrus.Fields{
				"user_name":        nameEntry.Text,
				"user_email":       emailEntry.Text,
				"user_description": descriptionEntry.Text,
				"error":            err.Error(),
			}).Info("Bug report generated")
		},
		fyne.CurrentApp().Driver().AllWindows()[0])

	reportDialog.Resize(fyne.NewSize(500, 600))
	reportDialog.Show()
}

// ShowInfo displays an information message
func (er *ErrorReport) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, fyne.CurrentApp().Driver().AllWindows()[0])
	er.logger.WithFields(logrus.Fields{
		"title":   title,
		"message": message,
	}).Info("Info dialog shown")
}

// ShowWarning displays a warning message
func (er *ErrorReport) ShowWarning(title, message string) {
	// Fyne doesn't have a built-in warning dialog, so we'll use a custom one
	content := container.NewVBox(
		widget.NewIcon(fyne.CurrentApp().Metadata().Icon),
		widget.NewLabel(message),
	)

	warningDialog := dialog.NewCustom(title, "OK", content, fyne.CurrentApp().Driver().AllWindows()[0])
	warningDialog.Show()

	er.logger.WithFields(logrus.Fields{
		"title":   title,
		"message": message,
	}).Warn("Warning dialog shown")
}

// LogDebugInfo logs debugging information about the current state
func (er *ErrorReport) LogDebugInfo(context string, data map[string]interface{}) {
	if er.debugMode {
		er.logger.WithFields(logrus.Fields(data)).Debug(context)
	}
}