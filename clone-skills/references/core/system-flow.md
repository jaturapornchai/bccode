# BC AI Cloud — System Flow

System overview: all modules and document flows between them.

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

    PO -->|"approved -> create PU"| PU
    PO -->|"approved -> create PP"| PP
    PU -->|"completed -> stock -"| PT
    PU -->|"completed -> stock +"| STOCK_DB
    PT -->|"completed -> stock -"| STOCK_DB
    PP -->|"accrue AP"| PI
    PP -->|"completed -> stock +"| STOCK_DB
    SO -->|"approved -> create SI"| SI
    SI -->|"completed -> stock -"| STOCK_DB
    SA -->|"completed -> stock +/-"| STOCK_DB
    TR -->|"completed -> transfer"| STOCK_DB

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
    PD -.->|"applied to"| PI["Purchase Invoice"]
    PC -.->|"applied to"| SI["Sales Invoice"]
```

---

## Document State Machine

All documents use the same state machine (PO, PI, SO, SI, SA, TR).

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

| State | Editable | Deletable | Printable |
|-------|----------|-----------|-----------|
| draft | Yes | Yes | Yes |
| pending | No | No | Yes |
| rejected | Yes | No | Yes |
| approved | No | No | Yes |
| completed | No | No | Yes |
| cancelled | No | No | Yes |
