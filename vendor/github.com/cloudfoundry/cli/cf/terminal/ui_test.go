package terminal_test

import (
	"io"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/trace/tracefakes"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	go_i18n "github.com/nicksnyder/go-i18n/i18n"

	. "github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var fakeLogger *tracefakes.FakePrinter
	BeforeEach(func() {
		fakeLogger = new(tracefakes.FakePrinter)
	})

	Describe("Printing message to stdout with PrintCapturingNoOutput", func() {
		It("prints strings without using the TeePrinter", func() {
			bucket := gbytes.NewBuffer()

			printer := NewTeePrinter()
			printer.SetOutputBucket(bucket)

			io_helpers.SimulateStdin("", func(reader io.Reader) {
				output := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, printer, fakeLogger)
					ui.PrintCapturingNoOutput("Hello")
				})

				Expect("Hello").To(Equal(strings.Join(output, "")))
				Expect(bucket.Contents()).To(HaveLen(0))
			})
		})
	})

	Describe("Printing message to stdout with Say", func() {
		It("prints strings", func() {
			io_helpers.SimulateStdin("", func(reader io.Reader) {
				output := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					ui.Say("Hello")
				})

				Expect("Hello").To(Equal(strings.Join(output, "")))
			})
		})

		It("prints formatted strings", func() {
			io_helpers.SimulateStdin("", func(reader io.Reader) {
				output := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					ui.Say("Hello %s", "World!")
				})

				Expect("Hello World!").To(Equal(strings.Join(output, "")))
			})
		})

		It("does not format strings when provided no args", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.Say("Hello %s World!") // whoops
			})

			Expect(strings.Join(output, "")).To(Equal("Hello %s World!"))
		})
	})

	Describe("Asking user for input", func() {
		It("allows string with whitespaces", func() {
			io_helpers.CaptureOutput(func() {
				io_helpers.SimulateStdin("foo bar\n", func(reader io.Reader) {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Ask("?")).To(Equal("foo bar"))
				})
			})
		})

		It("returns empty string if an error occured while reading string", func() {
			io_helpers.CaptureOutput(func() {
				io_helpers.SimulateStdin("string without expected delimiter", func(reader io.Reader) {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Ask("?")).To(Equal(""))
				})
			})
		})

		It("always outputs the prompt, even when output is disabled", func() {
			output := io_helpers.CaptureOutput(func() {
				io_helpers.SimulateStdin("things are great\n", func(reader io.Reader) {
					printer := NewTeePrinter()
					printer.DisableTerminalOutput(true)
					ui := NewUI(reader, printer, fakeLogger)
					ui.Ask("You like things?")
				})
			})
			Expect(strings.Join(output, "")).To(ContainSubstring("You like things?"))
		})
	})

	Describe("Confirming user input", func() {
		It("treats 'y' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("y\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Confirm("Hello World?")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})

		It("treats 'yes' as an affirmative confirmation when default language is not en_US", func() {
			oldLang := os.Getenv("LC_ALL")
			defer os.Setenv("LC_ALL", oldLang)

			oldT := i18n.T
			defer func() {
				i18n.T = oldT
			}()

			os.Setenv("LC_ALL", "fr_FR")

			config := configuration.NewRepositoryWithDefaults()
			i18n.T = i18n.Init(config)

			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Confirm("Hello World?")).To(BeTrue())
				})
				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Confirm("Hello World?")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})

		It("treats other input as a negative confirmation", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.Confirm("Hello World?")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})
	})

	Describe("Confirming deletion", func() {
		It("formats a nice output string with exactly one prompt", func() {
			io_helpers.SimulateStdin("y\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.ConfirmDelete("fizzbuzz", "bizzbump")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{
					"Really delete the fizzbuzz",
					"bizzbump",
					"?> ",
				}))
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"modelType modelName"}))
			})
		})

		It("treats other input as a negative confirmation and warns the user", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Delete cancelled"}))
			})
		})
	})

	Describe("Confirming deletion with associations", func() {
		It("warns the user that associated objects will also be deleted", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader, NewTeePrinter(), fakeLogger)
					Expect(ui.ConfirmDeleteWithAssociations("modelType", "modelName")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Delete cancelled"}))
			})
		})
	})

	Context("when user is not logged in", func() {
		var config coreconfig.Reader

		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("prompts the user to login", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.ShowConfiguration(config)
			})

			Expect(output).ToNot(ContainSubstrings([]string{"API endpoint:"}))
			Expect(output).To(ContainSubstrings([]string{"Not logged in", "Use", "log in"}))
		})
	})

	Context("when an api endpoint is set and the user logged in", func() {
		var config coreconfig.ReadWriter

		BeforeEach(func() {
			accessToken := coreconfig.TokenInfo{
				UserGuid: "my-user-guid",
				Username: "my-user",
				Email:    "my-user-email",
			}
			config = testconfig.NewRepositoryWithAccessToken(accessToken)
			config.SetApiEndpoint("https://test.example.org")
			config.SetApiVersion("☃☃☃")
		})

		Describe("tells the user what is set in the config", func() {
			var output []string

			JustBeforeEach(func() {
				output = io_helpers.CaptureOutput(func() {
					ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
					ui.ShowConfiguration(config)
				})
			})

			It("tells the user which api endpoint is set", func() {
				Expect(output).To(ContainSubstrings([]string{"API endpoint:", "https://test.example.org"}))
			})

			It("tells the user the api version", func() {
				Expect(output).To(ContainSubstrings([]string{"API version:", "☃☃☃"}))
			})

			It("tells the user which user is logged in", func() {
				Expect(output).To(ContainSubstrings([]string{"User:", "my-user-email"}))
			})

			Context("when an org is targeted", func() {
				BeforeEach(func() {
					config.SetOrganizationFields(models.OrganizationFields{
						Name: "org-name",
						Guid: "org-guid",
					})
				})

				It("tells the user which org is targeted", func() {
					Expect(output).To(ContainSubstrings([]string{"Org:", "org-name"}))
				})
			})

			Context("when a space is targeted", func() {
				BeforeEach(func() {
					config.SetSpaceFields(models.SpaceFields{
						Name: "my-space",
						Guid: "space-guid",
					})
				})

				It("tells the user which space is targeted", func() {
					Expect(output).To(ContainSubstrings([]string{"Space:", "my-space"}))
				})
			})
		})

		It("prompts the user to target an org and space when no org or space is targeted", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"}))
		})

		It("prompts the user to target an org when no org is targeted", func() {
			sf := models.SpaceFields{}
			sf.Guid = "guid"
			sf.Name = "name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "org", "targeted", "-o ORG"}))
		})

		It("prompts the user to target a space when no space is targeted", func() {
			of := models.OrganizationFields{}
			of.Guid = "of-guid"
			of.Name = "of-name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "space", "targeted", "-s SPACE"}))
		})
	})

	Describe("failing", func() {
		It("panics with a specific string", func() {
			io_helpers.CaptureOutput(func() {
				testassert.AssertPanic(QuietPanic, func() {
					NewUI(os.Stdin, NewTeePrinter(), fakeLogger).Failed("uh oh")
				})
			})
		})

		Context("when 'T' func is not initialized", func() {
			var t go_i18n.TranslateFunc
			BeforeEach(func() {
				t = i18n.T
				i18n.T = nil
			})

			AfterEach(func() {
				i18n.T = t
			})

			It("does not use 'T' func to translate", func() {
				io_helpers.CaptureOutput(func() {
					testassert.AssertPanic(QuietPanic, func() {
						NewUI(os.Stdin, NewTeePrinter(), fakeLogger).Failed("uh oh")
					})
				})
			})

			It("does not duplicate output if logger is set to stdout", func() {
				output := io_helpers.CaptureOutput(func() {
					testassert.AssertPanic(QuietPanic, func() {
						logger := trace.NewWriterPrinter(os.Stdout, true)
						NewUI(os.Stdin, NewTeePrinter(), logger).Failed("this should print only once")
					})
				})

				Expect(output).To(HaveLen(3))
				Expect(output[0]).To(Equal("FAILED"))
				Expect(output[1]).To(Equal("this should print only once"))
				Expect(output[2]).To(Equal(""))
			})
		})

		Context("when 'T' func is initialized", func() {
			It("does not duplicate output if logger is set to stdout", func() {
				output := io_helpers.CaptureOutput(func() {
					testassert.AssertPanic(QuietPanic, func() {
						logger := trace.NewWriterPrinter(os.Stdout, true)
						NewUI(os.Stdin, NewTeePrinter(), logger).Failed("this should print only once")
					})
				})

				Expect(output).To(HaveLen(3))
				Expect(output[0]).To(Equal("FAILED"))
				Expect(output[1]).To(Equal("this should print only once"))
				Expect(output[2]).To(Equal(""))
			})
		})
	})

	Describe("NotifyUpdateIfNeeded", func() {

		var (
			output []string
			config coreconfig.ReadWriter
		)

		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("Prints a notification to user if current version < min cli version", func() {
			config.SetMinCliVersion("6.0.0")
			config.SetMinRecommendedCliVersion("6.5.0")
			config.SetApiVersion("2.15.1")
			cf.Version = "5.0.0"
			output = io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.NotifyUpdateIfNeeded(config)
			})

			Expect(output).To(ContainSubstrings([]string{"Cloud Foundry API version",
				"requires CLI version 6.0.0",
				"You are currently on version 5.0.0",
				"To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			}))
		})

		It("Doesn't print a notification to user if current version >= min cli version", func() {
			config.SetMinCliVersion("6.0.0")
			config.SetMinRecommendedCliVersion("6.5.0")
			config.SetApiVersion("2.15.1")
			cf.Version = "6.0.0"
			output = io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin, NewTeePrinter(), fakeLogger)
				ui.NotifyUpdateIfNeeded(config)
			})

			Expect(output[0]).To(Equal(""))
		})
	})
})
