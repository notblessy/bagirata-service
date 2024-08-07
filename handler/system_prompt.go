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
- When there is tax or discount or Service charge or SC, they're not always use percentage, if the amount is greater than 100, get percentage from (tax amount / grand total from items +- otherPayments) * 100, otherwise get the exact value.
- Other payments type other than tax or discount & SC, get the exact value instead of percentage.
`
