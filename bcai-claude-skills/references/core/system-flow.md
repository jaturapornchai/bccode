# BC AI Cloud — System Flow

ภาพรวมระบบ BC AI Cloud: module ทั้งหมดและ document flow ระหว่างกัน

---

## System Overview

```mermaid
flowchart TD
    subgraph PURCHASE["Purchase"]
        PO["Purchase Order (PO)"]
        PU["Purchase Invoice (PU)"]
        PT["Purchase Return (PT)"]
        PP["Goods Receipt Note (PP)"]
        PI["Partial Purchase Invoice (PI)"]
    end

    subgraph SALES["Sales"]
        SO["Sales Order (SO)"]
        SI["Sales Invoice (SI)"]
    end

    subgraph STOCK["Stock Management"]
        SA["Stock Adjustment (SA)"]
        TR["Transfer (TR)"]
        STOCK_DB[("Stock Balance")]
    end

    subgraph PAYMENT["Payment"]
        PD["Advance Payment (PD)"]
        PDR["Payment Refund (PDR)"]
        PC["Deposit (PC)"]
        PCR["Deposit Refund (PCR)"]
    end

    subgraph GL["GL"]
        GL_POST["GL Posting (async)"]
    end

    PO -->|"approved → สร้าง PU"| PU
    PO -->|"approved → สร้าง PP"| PP
    PU -->|"completed → stock -"| PT
    PU -->|"completed → stock +"| STOCK_DB
    PT -->|"completed → stock -"| STOCK_DB
    PP -->|"ตั้งหนี้"| PI
    PP -->|"completed → stock +"| STOCK_DB
    SO -->|"approved → สร้าง SI"| SI
    SI -->|"completed → stock -"| STOCK_DB
    SA -->|"completed → stock ±"| STOCK_DB
    TR -->|"completed → โอนย้าย"| STOCK_DB

    PI --> GL_POST
    SI --> GL_POST
    SA --> GL_POST
    TR --> GL_POST
```

---

## Payment Flow

```mermaid
flowchart LR
    PD["Advance Payment\n(PD)"] -->|"refund"| PDR["Payment Refund\n(PDR)"]
    PC["Deposit\n(PC)"] -->|"refund"| PCR["Deposit Refund\n(PCR)"]
    PD -.->|"ใช้กับ"| PI["Purchase Invoice"]
    PC -.->|"ใช้กับ"| SI["Sales Invoice"]
```

---

## Document State Machine

ทุก document ใช้ state machine เดียวกัน (PO, PI, SO, SI, SA, TR)

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> pending : submit
    pending --> approved : approve
    pending --> rejected : reject
    rejected --> pending : resubmit
    approved --> completed : complete
    approved --> cancelled : cancel
    completed --> [*]
    cancelled --> [*]
```

| State | แก้ไขได้ | ลบได้ | พิมพ์ได้ |
|---|---|---|---|
| draft | yes | yes | yes |
| pending | no | no | yes |
| rejected | yes | no | yes |
| approved | no | no | yes |
| completed | no | no | yes |
| cancelled | no | no | yes |
