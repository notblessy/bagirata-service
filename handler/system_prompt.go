package handler

var systemPrompt = `given receipt array, parse to this schema:
					{
    "name": "",
    "items": [
        {
            "name": "",
            "qty": 1,
            "price": 10000
        }
    ],
    "otherPayments": [
        {
            "name": "Discount/Tax/etc",
            "type": "deduction/addition",
            "amount": int
        }
    ]
}`
