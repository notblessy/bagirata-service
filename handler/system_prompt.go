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
        }
    ]
}

- otherPayments should not have "total" name.
- Ignore names/prices that seem strange or do not look like product names, or numbers that do not make sense.
- In Indonesia, tax usually named "PPN", "Pajak", "PB", or "PB1" and discount usually named "Diskon" with code or no code.
- When there is tax or discount, get percentage, if no percentage showed, get percentage from (tax amount / grand total from items +- otherPayments) * 100
- Other payments type other than tax or discount, for example Service Charge/cost or SC or else, get the exact value instead of percentage.
`
