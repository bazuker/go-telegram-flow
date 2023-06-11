[![Go Report Card](https://goreportcard.com/badge/github.com/kisulken/go-telegram-flow)](https://goreportcard.com/report/github.com/kisulken/go-telegram-flow)

# go-telegram-flow [Deprecated]
Flow is a small framework for Telegram that is based on [this telegram bot package](https://github.com/tucnak/telebot) and [this locale package](https://github.com/tucnak/tr)

```Bash
go get -u github.com/kisulken/go-telegram-flow
```

With this library you can create dynamic menus & logical chains in Telegram just by defining the flow! Wow!

To see the full example check **_examples** directory
```Go
    // menu
	flow, err := menu.NewMenuFlow("flow1", b, "_examples/menu/lang", defaultLocale)
	if err != nil {
		panic(err)
	} 
	flow.GetRoot().
		Add("greetings", userPressGreeting).
		AddWith("order", userPress,
			flow.NewNode("pizza", userPress).
				Add("margarita", userOrderPizza).
				Add("pepperoni", userOrderPizza).
				Add("back", userPressBack), // traditional way of making back buttons
			flow.NewNode("sushi", userPress).
				Add("temaki", userOrderSushi).
				Add("nigiri", userOrderSushi).
				Add("sasazushi", userOrderSushi).
				Add("back", userPressBack),
			flow.NewBackNode("back"), // a short hand for making back buttons
		).
		Add("invoice", userPressInvoice).
		Add("language", userPressLanguage).GetFlow().Build("en").Build("ru")
```
```Go
    // chain
	flow, err = chain.NewChainFlow("flow1", b)
	if err != nil {
		panic(err)
	}

	flow.DefaultHandler(defaultResponse).GetRoot().
		Then("name", stageName, tb.OnText).
		Then("phone", stagePhone, tb.OnContact).
		Then("share_location", stageShareLocation, tb.OnText).
		Then("location", stageLocation, tb.OnLocation)
```

