# Manual Writing Standards

## Principle
Manuals are user-facing documentation. Write with enough detail that users understand without asking anyone.

## Rules

### 1. Describe Every Field
- **Field name** + meaning
- **Default value** (if any)
- **Options/value range**
- **Real usage example**
- **Impact** of incorrect settings

### 2. Cross-Reference with Source Code
- Check **Go backend model** → are all fields covered?
- Check **Flutter frontend** → what fields does the UI show?
- Check **MongoDB data** → what are actual values?
- If a field exists in source code but not in manual → **add it**

### 3. Manual Structure
- Clear sections grouped by feature
- Summary/overview at the top
- Tips/recommendations at the end
- Use icons/emoji for categorization (e.g., ⚙️ 📋 💡)

### 4. Language
- Write in **Thai**, easy to understand
- Parenthesize English for technical terms, e.g., "สกุลเงินหลัก (Base Currency)"
- Friendly, clear tone — avoid overly technical jargon

### 5. Website Manual Page Format (Next.js)
- Tailwind CSS + responsive design
- Breadcrumb navigation
- Sidebar for section jumping
- Screenshots/illustrations when available
- Update `lib/search-index.json` for chatbot discoverability

## Pre-Publish Checklist
- [ ] All backend/frontend fields covered
- [ ] Every field has description + example
- [ ] Default values match source code
- [ ] Thai text is clear and readable
- [ ] Build passes with no errors
