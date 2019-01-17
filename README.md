# go-telegram-flow
Menu flow is a paginated menu of inline buttons for Telegram

```Bash
go get -u github.com/kisulken/go-telegram-flow
```

With this library you can create dynamic menus in Telegram just by defining the flow! Wow!

<img src="https://drive.google.com/uc?id=174701OOF1wD6Eqs2u7K-EYfCfFlkvZTq&export=download" alt="" data-canonical-src="https://gyazo.com/eb5c5741b6a9a16c692170a41a49c858.png" width="270" height="480" />

This library is based on [this telegram bot package](https://github.com/tucnak/telebot) and [this locale package](https://github.com/tucnak/tr)

To see the full example check **_examples** directory
```Go
flow.Root.Add("greetings", userPressGreeting).
		AddWith("order", userPress,
			flow.New("pizza", userPress).
				Add("margarita", userOrderPizza).
				Add("pepperoni", userOrderPizza).Add("back", userPressBack),
			flow.New("sushi", userPress).
				Add("temaki", userOrderSushi).
				Add("nigiri", userOrderSushi).Add("sasazushi", userOrderSushi).
				Add("back", userPressBack),
			flow.New("back", userPressBack),
		).
		Add("invoice", userPressInvoice).
		Add("language", userPressLanguage).Flow.Build("en").Build("ru")
```
