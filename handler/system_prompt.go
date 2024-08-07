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
            "type": "deduction/addition/tax/discount",
            "amount": int
            "usePercentage": bool // true if percentage, false if exact value
        }
    ]
}

- otherPayments should not have "total" name.
- Ignore names/prices that seem strange or do not look like product names, or numbers that do not make sense.
- In Indonesia, tax usually named "PPN", "Pajak", "PB", or "PB1" and discount usually named "Diskon" with code or no code.
- Other payments is not always use percentage. If it is, use percentage as true, otherwise get the exact value instead of percentage.
`
