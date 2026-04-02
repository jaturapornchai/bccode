---
name: payment
description: Payments, cheques, deposits, QR, multi-currency — Payment System
user-invocable: true
---

# Payment System

## Overview
Manage receiving/making payments, cheques, deposits, QR payments, and multi-currency support.

## System Structure

### Payment
| Type | Backend Path | Purpose |
|------|-------------|---------|
| Make Payment | `transaction/paid/` | Record outgoing payments |
| Receive Payment | `transaction/pay/` | Record incoming payments |
| Details | `transaction/paymentdetail/` | Payment details |
| Methods | `backend/internal/paymentmaster/` | Payment method master data |

### Deposit
| Type | Backend Path | Purpose |
|------|-------------|---------|
| Place Deposit | `transaction/deposit/` | Customer places deposit |
| Refund Deposit | `transaction/depositrefund/` | Refund deposit |
| Deposit Record | `transaction/depositrecord/` | Deposit history |
| Receive Deposit | `transaction/receivedeposit/` | Receive deposit from customer |
| Refund Received | `transaction/receivedepositrefund/` | Refund received deposit |

### Advance Payment
| Type | Backend Path | Purpose |
|------|-------------|---------|
| Advance Payment | `transaction/advancepayment/` | Make advance payment |
| Advance Refund | `transaction/advancepaymentrefund/` | Refund advance payment |
| Paid Advance | `transaction/paidadvance/` | Record advance payment |
| Paid Advance Refund | `transaction/paidadvancerefund/` | Refund paid advance |

### Cheque — 8 Types
| Type | Backend Path | Purpose |
|------|-------------|---------|
| Cheque Deposit | `transaction/chequedeposit/` | Deposit cheque at bank |
| Cheque Pass | `transaction/chequepass/` | Cheque cleared |
| Cheque Change | `transaction/chequechange/` | Replace cheque |
| Cheque Return | `transaction/chequereturn/` | Cheque bounced/returned |
| Cheque Disqualified | `transaction/chequedisqualified/` | Cheque rejected |
| Cheque Renew | `transaction/chequerenew/` | Request new cheque |
| Payment Cheque Deposit | `transaction/chequepaymentdeposit/` | Deposit issued cheque |
| Payment Cheque Change | `transaction/chequepaymentchange/` | Replace issued cheque |

### Bank & QR Payment
| Section | File | Purpose |
|---------|------|---------|
| Bank | `backend/internal/payment/bankmaster/` | Bank master data |
| Book Bank | `backend/internal/payment/bookbank/` | Bank accounts |
| QR Payment | `backend/internal/payment/qrpayment/` | QR Code payment |
| Transfer | `transaction/banktransferrecord/` | Bank transfer records |
| Withdrawal | `transaction/withdrawalrecord/` | Withdrawal records |
| Credit Card | `transaction/creditcardwithdrawal/` | Credit card withdrawal |

### Currency
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/currency/` | Currency management |
| BLoC | `bloc/currency/` | State management |
| Screen | `screens/currency/` | Exchange rate settings |

## Payment Flow
```
Customer places order (Sale Invoice)
  |
Select payment method: Cash / Transfer / Cheque / QR / Credit Card
  |
Record payment received (Pay)
  |
Issue receipt (Receipt)

For cheques:
Receive cheque → Deposit cheque → Wait for clearing → Cheque passed
                                                    → Cheque bounced → Replace/Renew
```

## Supported Payment Methods
| Method | Code | Description |
|--------|------|-------------|
| Cash | CASH | Cash payment/receipt |
| Transfer | TRANSFER | Bank transfer |
| Cheque | CHEQUE | Bank cheque |
| QR Code | QR | PromptPay / QR Payment |
| Credit Card | CREDIT | Credit card swipe |
| Deposit | DEPOSIT | Deduct from deposit |
| Credit Term | CREDIT_TERM | Credit sales (with payment due date) |

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `trans_bloc` | General transactions |
| `transaction_paidpay_bloc` | Payment processing |
| `wallet_pay_bloc` | Wallet payment |
| `qr_bloc` | QR Payment |
| `point_transaction_bloc` | Loyalty points |

### Repositories
`trans_repository`, `transaction_paidpay_repository`, `qrpayment_repository`, `point_transaction_repository`, `cash_in_drawer_repository`

### Services
`transaction_calculator_service` — Calculates totals, VAT, discounts, withholding tax

### Models
`transaction_model`, `wallet_model`, `qr_model`, `point_transaction_model`, `cash_in_drawer_model`, `cash_drawer_models`, `wht_model`, `price_model`

## Important Notes
- Cheques have multiple statuses — must track the full lifecycle
- QR Payment requires bank API integration
- Foreign currency → CurrencyModel has `name` (string), not `names`
- Deposit is not the same as Advance Payment — different flows
- SMS transaction (`smsreceive/`) can receive bank deposit notifications via SMS
- `paymentmaster` = payment methods enabled by the shop
