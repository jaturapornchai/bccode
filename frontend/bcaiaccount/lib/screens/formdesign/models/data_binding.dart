/// Data binding field definition for form designer
/// ใช้เชื่อมต่อ element กับข้อมูลจริงจากระบบ (MongoDB)
class DataBindingField {
  final String key;
  final String label;
  final String labelEn;
  final String category;
  final String dataType; // string, number, date, currency, image
  final String? defaultFormat;
  final String? sampleValue;

  /// MongoDB field name (bson tag) — e.g. 'docno', 'docdatetime'
  final String? mongoField;

  /// MongoDB collection name — e.g. 'transactions', 'transactiondetails'
  final String? mongoCollection;

  /// Description of this field (for UI tooltip)
  final String? description;

  const DataBindingField({
    required this.key,
    required this.label,
    required this.labelEn,
    required this.category,
    required this.dataType,
    this.defaultFormat,
    this.sampleValue,
    this.mongoField,
    this.mongoCollection,
    this.description,
  });
}

/// Categories for data binding fields
class DataBindingCategory {
  final String id;
  final String label;
  final String labelEn;
  final List<DataBindingField> fields;

  const DataBindingCategory({
    required this.id,
    required this.label,
    required this.labelEn,
    required this.fields,
  });
}

/// Registry of all available data binding fields
/// ข้อมูลมาจาก MongoDB collection schemas
class DataBindingRegistry {
  static const List<DataBindingCategory> categories = [
    // Document
    DataBindingCategory(
      id: 'document',
      label: 'เอกสาร',
      labelEn: 'Document',
      fields: [
        DataBindingField(key: 'doc_number', label: 'เลขที่เอกสาร', labelEn: 'Document No.', category: 'document', dataType: 'string', sampleValue: 'INV-2568-0001', mongoField: 'docno', mongoCollection: 'transactions', description: 'เลขที่เอกสาร เช่น IV-001, PO-001'),
        DataBindingField(key: 'doc_date', label: 'วันที่เอกสาร', labelEn: 'Document Date', category: 'document', dataType: 'date', defaultFormat: 'dd/MM/yyyy', sampleValue: '13/03/2568', mongoField: 'docdatetime', mongoCollection: 'transactions', description: 'วันที่ออกเอกสาร'),
        DataBindingField(key: 'doc_date_be', label: 'วันที่ (พ.ศ.)', labelEn: 'Date (BE)', category: 'document', dataType: 'date', defaultFormat: 'dd MMMM พ.ศ. yyyy', sampleValue: '13 มีนาคม พ.ศ. 2569', mongoField: 'docdatetime', mongoCollection: 'transactions', description: 'วันที่แสดงเป็นปี พ.ศ.'),
        DataBindingField(key: 'doc_date_ce', label: 'วันที่ (ค.ศ.)', labelEn: 'Date (CE)', category: 'document', dataType: 'date', defaultFormat: 'dd/MM/yyyy', sampleValue: '13/03/2026', mongoField: 'docdatetime', mongoCollection: 'transactions', description: 'วันที่แสดงเป็นปี ค.ศ.'),
        DataBindingField(key: 'doc_ref', label: 'เลขที่อ้างอิง', labelEn: 'Reference No.', category: 'document', dataType: 'string', sampleValue: 'PO-2568-0042', mongoField: 'refno', mongoCollection: 'transactions', description: 'เลขที่อ้างอิงเอกสารต้นทาง'),
        DataBindingField(key: 'doc_type_name', label: 'ชื่อประเภทเอกสาร', labelEn: 'Document Type', category: 'document', dataType: 'string', sampleValue: 'ใบกำกับภาษี', mongoField: 'transflagname', mongoCollection: 'transactions', description: 'ชื่อประเภทเอกสาร เช่น ใบกำกับภาษี, ใบเสร็จรับเงิน'),
        DataBindingField(key: 'doc_due_date', label: 'วันครบกำหนด', labelEn: 'Due Date', category: 'document', dataType: 'date', defaultFormat: 'dd/MM/yyyy', sampleValue: '13/04/2568', mongoField: 'duedate', mongoCollection: 'transactions', description: 'วันครบกำหนดชำระเงิน'),
        DataBindingField(key: 'doc_remark', label: 'หมายเหตุ', labelEn: 'Remark', category: 'document', dataType: 'string', sampleValue: 'ชำระภายใน 30 วัน', mongoField: 'remark', mongoCollection: 'transactions', description: 'หมายเหตุเอกสาร'),
        DataBindingField(key: 'doc_credit_days', label: 'เครดิต (วัน)', labelEn: 'Credit Days', category: 'document', dataType: 'number', sampleValue: '30', mongoField: 'creditday', mongoCollection: 'transactions', description: 'จำนวนวันเครดิต'),
        DataBindingField(key: 'page_number', label: 'เลขหน้า', labelEn: 'Page No.', category: 'document', dataType: 'number', sampleValue: '1', description: 'เลขหน้าปัจจุบัน (คำนวณอัตโนมัติ)'),
        DataBindingField(key: 'page_count', label: 'จำนวนหน้า', labelEn: 'Total Pages', category: 'document', dataType: 'number', sampleValue: '1', description: 'จำนวนหน้าทั้งหมด (คำนวณอัตโนมัติ)'),
      ],
    ),
    // Company/Branch
    DataBindingCategory(
      id: 'company',
      label: 'บริษัท/สาขา',
      labelEn: 'Company/Branch',
      fields: [
        DataBindingField(key: 'company_name', label: 'ชื่อบริษัท', labelEn: 'Company Name', category: 'company', dataType: 'string', sampleValue: 'บริษัท ตัวอย่าง จำกัด', mongoField: 'names.name1', mongoCollection: 'shops', description: 'ชื่อบริษัท/ร้านค้า ภาษาไทย'),
        DataBindingField(key: 'company_name_en', label: 'ชื่อบริษัท (EN)', labelEn: 'Company Name (EN)', category: 'company', dataType: 'string', sampleValue: 'Example Co., Ltd.', mongoField: 'names.name2', mongoCollection: 'shops', description: 'ชื่อบริษัท/ร้านค้า ภาษาอังกฤษ'),
        DataBindingField(key: 'company_tax_id', label: 'เลขผู้เสียภาษี', labelEn: 'Tax ID', category: 'company', dataType: 'string', sampleValue: '0105500000001', mongoField: 'taxid', mongoCollection: 'shops', description: 'เลขประจำตัวผู้เสียภาษี 13 หลัก'),
        DataBindingField(key: 'company_address', label: 'ที่อยู่', labelEn: 'Address', category: 'company', dataType: 'string', sampleValue: '123 ถ.สุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพฯ 10110', mongoField: 'address', mongoCollection: 'shops', description: 'ที่อยู่สำนักงาน'),
        DataBindingField(key: 'company_branch', label: 'สำนักงาน/สาขา', labelEn: 'Branch', category: 'company', dataType: 'string', sampleValue: 'สำนักงานใหญ่', mongoField: 'branchname', mongoCollection: 'branches', description: 'ชื่อสาขา (สำนักงานใหญ่/สาขาที่ x)'),
        DataBindingField(key: 'company_branch_code', label: 'รหัสสาขา', labelEn: 'Branch Code', category: 'company', dataType: 'string', sampleValue: '00000', mongoField: 'branchcode', mongoCollection: 'branches', description: 'รหัสสาขา 5 หลัก (00000=สำนักงานใหญ่)'),
        DataBindingField(key: 'company_phone', label: 'โทรศัพท์', labelEn: 'Phone', category: 'company', dataType: 'string', sampleValue: '02-123-4567', mongoField: 'telephone', mongoCollection: 'shops', description: 'เบอร์โทรศัพท์บริษัท'),
        DataBindingField(key: 'company_fax', label: 'แฟกซ์', labelEn: 'Fax', category: 'company', dataType: 'string', sampleValue: '02-123-4568', mongoField: 'faxno', mongoCollection: 'shops', description: 'เบอร์แฟกซ์'),
        DataBindingField(key: 'company_email', label: 'อีเมล', labelEn: 'Email', category: 'company', dataType: 'string', sampleValue: 'info@example.co.th', mongoField: 'email', mongoCollection: 'shops', description: 'อีเมลบริษัท'),
        DataBindingField(key: 'company_website', label: 'เว็บไซต์', labelEn: 'Website', category: 'company', dataType: 'string', sampleValue: 'www.example.co.th', mongoField: 'website', mongoCollection: 'shops', description: 'เว็บไซต์บริษัท'),
        DataBindingField(key: 'company_logo', label: 'โลโก้', labelEn: 'Logo', category: 'company', dataType: 'image', sampleValue: '', mongoField: 'logoimage', mongoCollection: 'shops', description: 'รูปโลโก้บริษัท'),
      ],
    ),
    // Contact (Customer/Supplier)
    DataBindingCategory(
      id: 'contact',
      label: 'ลูกค้า/ผู้จำหน่าย',
      labelEn: 'Customer/Supplier',
      fields: [
        DataBindingField(key: 'contact_code', label: 'รหัส', labelEn: 'Code', category: 'contact', dataType: 'string', sampleValue: 'C-0001', mongoField: 'custcode', mongoCollection: 'transactions', description: 'รหัสลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_name', label: 'ชื่อ', labelEn: 'Name', category: 'contact', dataType: 'string', sampleValue: 'บริษัท ลูกค้า จำกัด', mongoField: 'custnames.name1', mongoCollection: 'transactions', description: 'ชื่อลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_tax_id', label: 'เลขผู้เสียภาษี', labelEn: 'Tax ID', category: 'contact', dataType: 'string', sampleValue: '0105500000002', mongoField: 'custtaxid', mongoCollection: 'transactions', description: 'เลขประจำตัวผู้เสียภาษี ลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_address', label: 'ที่อยู่', labelEn: 'Address', category: 'contact', dataType: 'string', sampleValue: '456 ถ.พระราม 4 กรุงเทพฯ 10500', mongoField: 'custaddress', mongoCollection: 'transactions', description: 'ที่อยู่ลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_branch', label: 'สาขา', labelEn: 'Branch', category: 'contact', dataType: 'string', sampleValue: 'สำนักงานใหญ่', mongoField: 'custbranchname', mongoCollection: 'transactions', description: 'สาขาลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_phone', label: 'โทรศัพท์', labelEn: 'Phone', category: 'contact', dataType: 'string', sampleValue: '02-987-6543', mongoField: 'custtelephone', mongoCollection: 'transactions', description: 'เบอร์โทรศัพท์ลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_email', label: 'อีเมล', labelEn: 'Email', category: 'contact', dataType: 'string', sampleValue: 'customer@example.com', mongoField: 'custemail', mongoCollection: 'transactions', description: 'อีเมลลูกค้า/ผู้จำหน่าย'),
        DataBindingField(key: 'contact_attention', label: 'ผู้ติดต่อ', labelEn: 'Attention', category: 'contact', dataType: 'string', sampleValue: 'คุณสมชาย', mongoField: 'contactname', mongoCollection: 'transactions', description: 'ชื่อผู้ติดต่อ'),
      ],
    ),
    // Detail Line (repeating)
    DataBindingCategory(
      id: 'detail',
      label: 'รายการสินค้า',
      labelEn: 'Line Items',
      fields: [
        DataBindingField(key: 'line_number', label: 'ลำดับ', labelEn: 'No.', category: 'detail', dataType: 'number', sampleValue: '1', mongoField: 'linenumber', mongoCollection: 'transactiondetails', description: 'ลำดับรายการ'),
        DataBindingField(key: 'item_code', label: 'รหัสสินค้า', labelEn: 'Item Code', category: 'detail', dataType: 'string', sampleValue: 'P-001', mongoField: 'itemcode', mongoCollection: 'transactiondetails', description: 'รหัสสินค้า'),
        DataBindingField(key: 'item_barcode', label: 'บาร์โค้ด', labelEn: 'Barcode', category: 'detail', dataType: 'string', sampleValue: '8850000000001', mongoField: 'barcode', mongoCollection: 'transactiondetails', description: 'บาร์โค้ดสินค้า'),
        DataBindingField(key: 'item_name', label: 'ชื่อสินค้า/บริการ', labelEn: 'Description', category: 'detail', dataType: 'string', sampleValue: 'สินค้าตัวอย่าง A', mongoField: 'itemnames.name1', mongoCollection: 'transactiondetails', description: 'ชื่อสินค้า/บริการ'),
        DataBindingField(key: 'item_qty', label: 'จำนวน', labelEn: 'Qty', category: 'detail', dataType: 'number', defaultFormat: '#,##0', sampleValue: '10', mongoField: 'qty', mongoCollection: 'transactiondetails', description: 'จำนวน'),
        DataBindingField(key: 'item_unit', label: 'หน่วย', labelEn: 'Unit', category: 'detail', dataType: 'string', sampleValue: 'ชิ้น', mongoField: 'unitcode', mongoCollection: 'transactiondetails', description: 'รหัสหน่วยนับ'),
        DataBindingField(key: 'item_price', label: 'ราคาต่อหน่วย', labelEn: 'Unit Price', category: 'detail', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '1,500.00', mongoField: 'price', mongoCollection: 'transactiondetails', description: 'ราคาต่อหน่วย'),
        DataBindingField(key: 'item_discount', label: 'ส่วนลด', labelEn: 'Discount', category: 'detail', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '150.00', mongoField: 'discamount', mongoCollection: 'transactiondetails', description: 'จำนวนเงินส่วนลดรายการ'),
        DataBindingField(key: 'item_amount', label: 'จำนวนเงิน', labelEn: 'Amount', category: 'detail', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '14,850.00', mongoField: 'sumamount', mongoCollection: 'transactiondetails', description: 'จำนวนเงินรวม (qty × price - discount)'),
        DataBindingField(key: 'item_remark', label: 'หมายเหตุ', labelEn: 'Remark', category: 'detail', dataType: 'string', sampleValue: '', mongoField: 'remark', mongoCollection: 'transactiondetails', description: 'หมายเหตุรายการ'),
      ],
    ),
    // Financial
    DataBindingCategory(
      id: 'financial',
      label: 'สรุปยอดเงิน',
      labelEn: 'Financial Summary',
      fields: [
        DataBindingField(key: 'subtotal', label: 'รวมเงิน', labelEn: 'Subtotal', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '14,850.00', mongoField: 'totalvalue', mongoCollection: 'transactions', description: 'รวมมูลค่าสินค้า/บริการก่อนหักส่วนลด'),
        DataBindingField(key: 'discount_total', label: 'ส่วนลดรวม', labelEn: 'Total Discount', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '0.00', mongoField: 'totaldiscountamount', mongoCollection: 'transactions', description: 'ส่วนลดรวมทั้งเอกสาร'),
        DataBindingField(key: 'amount_before_vat', label: 'จำนวนเงินก่อนภาษี', labelEn: 'Amount Before VAT', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '13,878.50', mongoField: 'totalbeforevat', mongoCollection: 'transactions', description: 'จำนวนเงินก่อนภาษีมูลค่าเพิ่ม'),
        DataBindingField(key: 'vat_rate', label: 'อัตราภาษี (%)', labelEn: 'VAT Rate', category: 'financial', dataType: 'number', defaultFormat: '#,##0', sampleValue: '7', mongoField: 'vatrate', mongoCollection: 'transactions', description: 'อัตราภาษีมูลค่าเพิ่ม (%)'),
        DataBindingField(key: 'vat_amount', label: 'ภาษีมูลค่าเพิ่ม', labelEn: 'VAT Amount', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '971.50', mongoField: 'totalvatvalue', mongoCollection: 'transactions', description: 'จำนวนภาษีมูลค่าเพิ่ม'),
        DataBindingField(key: 'total_amount', label: 'จำนวนเงินรวมทั้งสิ้น', labelEn: 'Grand Total', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '14,850.00', mongoField: 'totalamount', mongoCollection: 'transactions', description: 'จำนวนเงินรวมทั้งสิ้น (รวม VAT)'),
        DataBindingField(key: 'total_text_th', label: 'จำนวนเงิน (ตัวอักษร)', labelEn: 'Amount in Words (TH)', category: 'financial', dataType: 'string', sampleValue: 'หนึ่งหมื่นสี่พันแปดร้อยห้าสิบบาทถ้วน', mongoField: 'totalamounttext', mongoCollection: 'transactions', description: 'จำนวนเงินเป็นตัวอักษรภาษาไทย'),
        DataBindingField(key: 'total_text_en', label: 'Amount in Words (EN)', labelEn: 'Amount in Words (EN)', category: 'financial', dataType: 'string', sampleValue: 'Fourteen Thousand Eight Hundred and Fifty Baht Only', mongoField: 'totalamounttexten', mongoCollection: 'transactions', description: 'จำนวนเงินเป็นตัวอักษรภาษาอังกฤษ'),
        DataBindingField(key: 'withholding_tax', label: 'ภาษีหัก ณ ที่จ่าย', labelEn: 'Withholding Tax', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '445.50', mongoField: 'totalwhtamount', mongoCollection: 'transactions', description: 'ภาษีหัก ณ ที่จ่าย'),
        DataBindingField(key: 'net_amount', label: 'ยอดสุทธิ', labelEn: 'Net Amount', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '14,404.50', mongoField: 'totalnetamount', mongoCollection: 'transactions', description: 'ยอดสุทธิ (Grand Total - WHT)'),
        DataBindingField(key: 'deposit_amount', label: 'เงินมัดจำ', labelEn: 'Deposit', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '0.00', mongoField: 'depositamount', mongoCollection: 'transactions', description: 'เงินมัดจำ/เงินจอง'),
        DataBindingField(key: 'balance_amount', label: 'ยอดค้างชำระ', labelEn: 'Balance', category: 'financial', dataType: 'currency', defaultFormat: '#,##0.00', sampleValue: '14,404.50', mongoField: 'balanceamount', mongoCollection: 'transactions', description: 'ยอดค้างชำระ'),
      ],
    ),
    // Signature
    DataBindingCategory(
      id: 'signature',
      label: 'ลายเซ็น/ผู้ลงนาม',
      labelEn: 'Signature',
      fields: [
        DataBindingField(key: 'prepared_by', label: 'ผู้จัดทำ', labelEn: 'Prepared By', category: 'signature', dataType: 'string', sampleValue: 'คุณสมหญิง', mongoField: 'creatorname', mongoCollection: 'transactions', description: 'ชื่อผู้จัดทำเอกสาร'),
        DataBindingField(key: 'approved_by', label: 'ผู้อนุมัติ', labelEn: 'Approved By', category: 'signature', dataType: 'string', sampleValue: 'คุณสมชาย', mongoField: 'approvername', mongoCollection: 'transactions', description: 'ชื่อผู้อนุมัติเอกสาร'),
        DataBindingField(key: 'received_by', label: 'ผู้รับ', labelEn: 'Received By', category: 'signature', dataType: 'string', sampleValue: '', description: 'ชื่อผู้รับสินค้า/เอกสาร (กรอกตอนพิมพ์)'),
        DataBindingField(key: 'delivered_by', label: 'ผู้ส่ง', labelEn: 'Delivered By', category: 'signature', dataType: 'string', sampleValue: '', mongoField: 'deliveryname', mongoCollection: 'transactions', description: 'ชื่อผู้จัดส่ง'),
        DataBindingField(key: 'signature_date', label: 'วันที่ลงนาม', labelEn: 'Signature Date', category: 'signature', dataType: 'date', defaultFormat: 'dd/MM/yyyy', sampleValue: '13/03/2568', description: 'วันที่ลงนาม (ใช้วันที่พิมพ์)'),
      ],
    ),
    // Payment
    DataBindingCategory(
      id: 'payment',
      label: 'การชำระเงิน',
      labelEn: 'Payment',
      fields: [
        DataBindingField(key: 'payment_method', label: 'วิธีชำระ', labelEn: 'Payment Method', category: 'payment', dataType: 'string', sampleValue: 'โอนเงิน', mongoField: 'paymenttype', mongoCollection: 'transactions', description: 'วิธีการชำระเงิน เช่น เงินสด, โอน, บัตรเครดิต'),
        DataBindingField(key: 'bank_name', label: 'ธนาคาร', labelEn: 'Bank', category: 'payment', dataType: 'string', sampleValue: 'ธ.กสิกรไทย', mongoField: 'bankname', mongoCollection: 'bookbanks', description: 'ชื่อธนาคาร'),
        DataBindingField(key: 'bank_account', label: 'เลขบัญชี', labelEn: 'Account No.', category: 'payment', dataType: 'string', sampleValue: '123-4-56789-0', mongoField: 'accountno', mongoCollection: 'bookbanks', description: 'เลขที่บัญชีธนาคาร'),
        DataBindingField(key: 'bank_account_name', label: 'ชื่อบัญชี', labelEn: 'Account Name', category: 'payment', dataType: 'string', sampleValue: 'บจก.ตัวอย่าง', mongoField: 'accountname', mongoCollection: 'bookbanks', description: 'ชื่อบัญชีธนาคาร'),
        DataBindingField(key: 'payment_qr', label: 'QR ชำระเงิน', labelEn: 'Payment QR', category: 'payment', dataType: 'image', sampleValue: '', description: 'QR Code สำหรับชำระเงิน (PromptPay/อื่นๆ)'),
      ],
    ),
    // Delivery
    DataBindingCategory(
      id: 'delivery',
      label: 'การจัดส่ง',
      labelEn: 'Delivery',
      fields: [
        DataBindingField(key: 'delivery_address', label: 'ที่อยู่จัดส่ง', labelEn: 'Delivery Address', category: 'delivery', dataType: 'string', sampleValue: '789 ถ.รัชดา กรุงเทพฯ 10310', mongoField: 'deliveryaddress', mongoCollection: 'transactions', description: 'ที่อยู่จัดส่งสินค้า'),
        DataBindingField(key: 'delivery_date', label: 'วันที่จัดส่ง', labelEn: 'Delivery Date', category: 'delivery', dataType: 'date', defaultFormat: 'dd/MM/yyyy', sampleValue: '15/03/2568', mongoField: 'deliverydate', mongoCollection: 'transactions', description: 'วันที่จัดส่งสินค้า'),
        DataBindingField(key: 'shipping_method', label: 'วิธีจัดส่ง', labelEn: 'Shipping Method', category: 'delivery', dataType: 'string', sampleValue: 'Kerry Express', mongoField: 'transportcode', mongoCollection: 'transactions', description: 'รหัสวิธีจัดส่ง/ขนส่ง'),
        DataBindingField(key: 'tracking_number', label: 'เลขพัสดุ', labelEn: 'Tracking No.', category: 'delivery', dataType: 'string', sampleValue: 'TH12345678', mongoField: 'trackingno', mongoCollection: 'transactions', description: 'เลขพัสดุ/เลข tracking'),
      ],
    ),
  ];

  /// Find a field by key
  static DataBindingField? findField(String key) {
    for (final cat in categories) {
      for (final field in cat.fields) {
        if (field.key == key) return field;
      }
    }
    return null;
  }

  /// Get sample value for a key
  static String getSampleValue(String key) {
    return findField(key)?.sampleValue ?? '{$key}';
  }

  /// Get all fields as a flat list
  static List<DataBindingField> get allFields {
    return categories.expand((c) => c.fields).toList();
  }
}
