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
}

- otherPayments should not have "total" name.
- Ignore names/prices that seem strange or do not look like product names, or numbers that do not make sense. 
`
