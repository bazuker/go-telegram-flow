# go-telegram-flow
Menu flow is a paginated menu of inline buttons for Telegram

With this library you can create dynamic menus in Telegram just by defining the flow! Wow!

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