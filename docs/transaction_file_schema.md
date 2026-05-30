# Transaction File Schema

This document defines the current hard-coded schema used to validate uploaded
Excel and CSV transaction files. It corresponds to schema version
`transaction-file-v1` in code and matches the source workbook format currently
used by the team.

## Validation model

- Validation happens after the file has been parsed into a tabular structure.
- Header row is treated as spreadsheet row `1`.
- The first data row is treated as spreadsheet row `2`.
- Validation errors include both the row number and the column reference.

## Expected columns

| # | Column | Required | Type | Allowed values | Description |
|---|---|---|---|---|---|
| 1 | `No. of Transactions` | Yes | `integer` | — | Count of transaction rows |
| 2 | `Goods Description` | Yes | `text` | — | Goods description as provided by client banks |
| 3 | `Goods Classification (Sector)` | Yes | `text` | — | Goods classification sector |
| 4 | `Applicant (CG/RPA) or Sub-Borrower (RCF) Country` | Yes | `text` | — | Country of the applicant or sub-borrower |
| 5 | `Beneficiary Country` | No | `text` | — | Country of the beneficiary when supplied |
| 6 | `Source` | Yes | `text` | — | Source country/origin column from the workbook |
| 7 | `Destination` | Yes | `text` | — | Destination country column from the workbook |
| 8 | `Tenor > 1 year` | Yes | `text` | — | Tenor classification from the workbook |
| 9 | `E&S Category` | No | `text` | — | Environmental and social category when supplied |
| 10 | `PA Alignment` | Yes | `text` | — | Paris alignment classification |

## Example header row

```text
No. of Transactions,Goods Description,Goods Classification (Sector),Applicant (CG/RPA) or Sub-Borrower (RCF) Country,Beneficiary Country,Source,Destination,Tenor > 1 year,E&S Category,PA Alignment
```

## Example values

- `No. of Transactions`: `1`
- `Goods Description`: `ARTICLES, ACCESSORIES AND COMPLEMENTS OF CLOTHING`
- `Goods Classification (Sector)`: `Consumer Goods`
- `Source`: `SPAIN`
- `Destination`: `ARMENIA`

## Notes for partner banks

- Column names should match the schema exactly.
- Required columns must be present in the uploaded file.
- Required cells must be populated on each data row.
- Numeric columns must use plain numeric values that parse as integers.

## Supported upload formats

- `.csv`
- `.xls`
- `.xlsx`

Files are parsed into the tabular structure above before validation. If any row
fails validation, the upload is rejected and the API returns structured
row/column errors instead of partially ingesting the file.
