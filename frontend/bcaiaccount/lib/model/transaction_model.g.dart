// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'transaction_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TransactionModel _$TransactionModelFromJson(Map<String, dynamic> json) =>
    TransactionModel(
        shopid: json['shopid'] as String?,
        guidref: json['guidref'] as String,
        docno: json['docno'] as String,
        docdatetime: json['docdatetime'] as String,
        docdatelocal: json['docdatelocal'] as String?,
        doctimelocal: json['doctimelocal'] as String?,
        timezone: json['timezone'] as String?,
        docrefno: json['docrefno'] as String,
        docrefdate: json['docrefdate'] as String,
        docreftype: (json['docreftype'] as num).toInt(),
        doctype: (json['doctype'] as num).toInt(),
        vattype: (json['vattype'] as num).toInt(),
        custcode: json['custcode'] as String,
        salecode: json['salecode'] as String,
        salename: json['salename'] as String,
        discountword: json['discountword'] as String,
        totalcost: (json['totalcost'] as num).toDouble(),
        totalvalue: (json['totalvalue'] as num).toDouble(),
        totaldiscount: (json['totaldiscount'] as num).toDouble(),
        totalvatvalue: (json['totalvatvalue'] as num).toDouble(),
        totalbeforevat: (json['totalbeforevat'] as num).toDouble(),
        totalaftervat: (json['totalaftervat'] as num).toDouble(),
        totalexceptvat: (json['totalexceptvat'] as num).toDouble(),
        totalamount: (json['totalamount'] as num).toDouble(),
        cashiercode: json['cashiercode'] as String,
        posid: json['posid'] as String,
        membercode: json['membercode'] as String,
        vatrate: (json['vatrate'] as num).toDouble(),
        status: (json['status'] as num).toInt(),
        isref: json['isref'] as bool?,
        islocked: json['islocked'] as bool?,
        isclosed: json['isclosed'] as bool?,
        iscomparedsuccess: (json['iscomparedsuccess'] as num?)?.toInt(),
        isclosedmanual: json['isclosedmanual'] as bool?,
        closedmanualByCode: json['closedmanualByCode'] as String?,
        closedmanualByName: json['closedmanualByName'] as String?,
        closedmanualAt: json['closedmanualAt'] as String?,
        closedmanualReason: json['closedmanualReason'] as String?,
        detailcount: (json['detailcount'] as num?)?.toInt(),
        details: (json['details'] as List<dynamic>?)
            ?.map(
              (e) => TransactionDetailModel.fromJson(e as Map<String, dynamic>),
            )
            .toList(),
        inquirytype: (json['inquirytype'] as num).toInt(),
        taxdocno: json['taxdocno'] as String,
        taxdocdate: json['taxdocdate'] as String,
        transflag: (json['transflag'] as num).toInt(),
        iscancel: json['iscancel'] as bool,
        ismanualamount: json['ismanualamount'] as bool,
        paymentdetail: json['paymentdetail'] == null
            ? null
            : TransactionPayModel.fromJson(
                json['paymentdetail'] as Map<String, dynamic>,
              ),
        custnames: (json['custnames'] as List<dynamic>?)
            ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
            .toList(),
        guidfixed: json['guidfixed'] as String?,
        description: json['description'] as String?,
        paymentdetailraw: json['paymentdetailraw'] as String?,
        paymentStructs: (json['paymentStructs'] as List<dynamic>?)
            ?.map((e) => BillPayStruct.fromJson(e as Map<String, dynamic>))
            .toList(),
        paycashamount: (json['paycashamount'] as num?)?.toDouble(),
        billtaxtype: (json['billtaxtype'] as num?)?.toInt(),
        canceltime: json['canceltime'] as String?,
        canceldatetime: json['canceldatetime'] as String?,
        cancelusercode: json['cancelusercode'] as String?,
        cancelusername: json['cancelusername'] as String?,
        canceldescription: json['canceldescription'] as String?,
        cancelreason: json['cancelreason'] as String?,
        creatorcode: json['creator_code'] as String?,
        creatorname: json['creator_name'] as String?,
        createdat: json['created_at'] as String?,
        modifiercode: json['modifier_code'] as String?,
        modifiername: json['modifier_name'] as String?,
        modifiedat: json['modified_at'] as String?,
        fullvataddress: json['fullvataddress'] as String?,
        fullvatbranchnumber: json['fullvatbranchnumber'] as String?,
        fullvatname: json['fullvatname'] as String?,
        fullvatdocnumber: json['fullvatdocnumber'] as String?,
        fullvattaxid: json['fullvattaxid'] as String?,
        fullvatprint: json['fullvatprint'] as bool?,
        isvatregister: json['isvatregister'] as bool?,
        printcopybilldatetime: (json['printcopybilldatetime'] as List<dynamic>?)
            ?.map((e) => e as String)
            .toList(),
        tablenumber: json['tablenumber'] as String?,
        tableopendatetime: json['tableopendatetime'] as String?,
        tableclosedatetime: json['tableclosedatetime'] as String?,
        mancount: (json['mancount'] as num?)?.toInt(),
        womancount: (json['womancount'] as num?)?.toInt(),
        childcount: (json['childcount'] as num?)?.toInt(),
        tableallacratemode: json['tableallacratemode'] as bool?,
        buffetcode: json['buffetcode'] as String?,
        customertelephone: json['customertelephone'] as String?,
        totalqty: (json['totalqty'] as num?)?.toDouble(),
        totaldiscountvatamount: (json['totaldiscountvatamount'] as num?)
            ?.toDouble(),
        totaldiscountexceptvatamount:
            (json['totaldiscountexceptvatamount'] as num?)?.toDouble(),
        cashiername: json['cashiername'] as String?,
        paycashchange: (json['paycashchange'] as num?)?.toDouble(),
        sumqrcode: (json['sumqrcode'] as num?)?.toDouble(),
        sumcreditcard: (json['sumcreditcard'] as num?)?.toDouble(),
        summoneytransfer: (json['summoneytransfer'] as num?)?.toDouble(),
        sumcheque: (json['sumcheque'] as num?)?.toDouble(),
        sumcoupon: (json['sumcoupon'] as num?)?.toDouble(),
        detaildiscountformula: json['detaildiscountformula'] as String?,
        detailtotalamount: (json['detailtotalamount'] as num?)?.toDouble(),
        detailtotaldiscount: (json['detailtotaldiscount'] as num?)?.toDouble(),
        roundamount: (json['roundamount'] as num?)?.toDouble(),
        totalamountafterdiscount: (json['totalamountafterdiscount'] as num?)
            ?.toDouble(),
        detailtotalamountbeforediscount:
            (json['detailtotalamountbeforediscount'] as num?)?.toDouble(),
        sumcredit: (json['sumcredit'] as num?)?.toDouble(),
        salechannelcode: json['salechannelcode'] as String?,
        salechannelgp: (json['salechannelgp'] as num?)?.toDouble(),
        salechannelgptype: (json['salechannelgptype'] as num?)?.toInt(),
        takeaway: (json['takeaway'] as num?)?.toInt(),
        branch: json['branch'] == null
            ? null
            : BranchModel.fromJson(json['branch'] as Map<String, dynamic>),
        reftotaloriginal: (json['reftotaloriginal'] as num?)?.toDouble(),
        reftotalcorrect: (json['reftotalcorrect'] as num?)?.toDouble(),
        reftotaldiff: (json['reftotaldiff'] as num?)?.toDouble(),
        slipurl: json['slipurl'] as String?,
        ispos: json['ispos'] as bool?,
        isbom: json['isbom'] as bool?,
        isdelivery: json['isdelivery'] as bool?,
        deliveryamount: (json['deliveryamount'] as num?)?.toDouble(),
        istransport: json['istransport'] as bool?,
        transportcode: json['transportcode'] as String?,
        transportamount: (json['transportamount'] as num?)?.toDouble(),
        getpoint: (json['getpoint'] as num?)?.toInt(),
        usepoint: (json['usepoint'] as num?)?.toInt(),
        pointdiscountamount: (json['pointdiscountamount'] as num?)?.toInt(),
        paypointamount: (json['paypointamount'] as num?)?.toInt(),
        pointscode: json['pointscode'] as String?,
        coupons: (json['coupons'] as List<dynamic>?)
            ?.map((e) => CouponModel.fromJson(e as Map<String, dynamic>))
            .toList(),
        totalcouponamount: (json['totalcouponamount'] as num?)?.toDouble(),
        coupondiscountamount: (json['coupondiscountamount'] as num?)
            ?.toDouble(),
        couponcashamount: (json['couponcashamount'] as num?)?.toDouble(),
        docreferences: (json['docreferences'] as List<dynamic>?)
            ?.map((e) => DocReferenceModel.fromJson(e as Map<String, dynamic>))
            .toList(),
        advancepaymentdocs: (json['advancepaymentdocs'] as List<dynamic>?)
            ?.map(
              (e) => AdvancePaymentDocModel.fromJson(e as Map<String, dynamic>),
            )
            .toList(),
        sumadvancepayment: (json['sumadvancepayment'] as num?)?.toDouble(),
        sumdeposit: (json['sumdeposit'] as num?)?.toDouble(),
        depositdocs: (json['depositdocs'] as List<dynamic>?)
            ?.map((e) => DepositDocModel.fromJson(e as Map<String, dynamic>))
            .toList(),
        purchasetypecode: json['purchasetypecode'] as String?,
        purchasetypenames: (json['purchasetypenames'] as List<dynamic>?)
            ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
            .toList(),
        creditdays: (json['creditdays'] as num?)?.toInt(),
        duedate: json['duedate'] as String?,
      )
      ..currency = json['currency'] as String?
      ..currencysymbol = json['currencysymbol'] as String?
      ..exchangerate = (json['exchangerate'] as num?)?.toDouble()
      ..totalamountDoc = (json['totalamount_doc'] as num?)?.toDouble()
      ..totalvalueDoc = (json['totalvalue_doc'] as num?)?.toDouble()
      ..totaldiscountDoc = (json['totaldiscount_doc'] as num?)?.toDouble()
      ..totalvatvalueDoc = (json['totalvatvalue_doc'] as num?)?.toDouble()
      ..totalbeforevatDoc = (json['totalbeforevat_doc'] as num?)?.toDouble()
      ..totalaftervatDoc = (json['totalaftervat_doc'] as num?)?.toDouble()
      ..docCurrency = json['doc_currency'] as String?
      ..docCurrencySymbol = json['doc_currencysymbol'] as String?
      ..whtEntries = (json['wht_entries'] as List<dynamic>?)
          ?.map((e) => WHTEntryModel.fromJson(e as Map<String, dynamic>))
          .toList()
      ..isdelete = json['isdelete'] as bool?;

Map<String, dynamic> _$TransactionModelToJson(
  TransactionModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'shopid': instance.shopid,
  'guidref': instance.guidref,
  'docno': instance.docno,
  'docdatetime': instance.docdatetime,
  'docdatelocal': instance.docdatelocal,
  'doctimelocal': instance.doctimelocal,
  'timezone': instance.timezone,
  'docrefno': instance.docrefno,
  'docrefdate': instance.docrefdate,
  'docreftype': instance.docreftype,
  'doctype': instance.doctype,
  'vattype': instance.vattype,
  'inquirytype': instance.inquirytype,
  'transflag': instance.transflag,
  'custcode': instance.custcode,
  'custnames': instance.custnames?.map((e) => e.toJson()).toList(),
  'salecode': instance.salecode,
  'salename': instance.salename,
  'discountword': instance.discountword,
  'totalcost': instance.totalcost,
  'totalvalue': instance.totalvalue,
  'totaldiscount': instance.totaldiscount,
  'totalvatvalue': instance.totalvatvalue,
  'totalbeforevat': instance.totalbeforevat,
  'totalaftervat': instance.totalaftervat,
  'totalexceptvat': instance.totalexceptvat,
  'totalamount': instance.totalamount,
  'cashiercode': instance.cashiercode,
  'posid': instance.posid,
  'membercode': instance.membercode,
  'isref': instance.isref,
  'islocked': instance.islocked,
  'isclosed': instance.isclosed,
  'iscomparedsuccess': instance.iscomparedsuccess,
  'isclosedmanual': instance.isclosedmanual,
  'closedmanualByCode': instance.closedmanualByCode,
  'closedmanualByName': instance.closedmanualByName,
  'closedmanualAt': instance.closedmanualAt,
  'closedmanualReason': instance.closedmanualReason,
  'detailcount': instance.detailcount,
  'ismanualamount': instance.ismanualamount,
  'taxdocno': instance.taxdocno,
  'taxdocdate': instance.taxdocdate,
  'vatrate': instance.vatrate,
  'iscancel': instance.iscancel,
  'status': instance.status,
  'details': instance.details?.map((e) => e.toJson()).toList(),
  'paymentdetail': instance.paymentdetail?.toJson(),
  'paymentdetailraw': instance.paymentdetailraw,
  'paymentStructs': instance.paymentStructs?.map((e) => e.toJson()).toList(),
  'paycashamount': instance.paycashamount,
  'description': instance.description,
  'billtaxtype': instance.billtaxtype,
  'canceltime': instance.canceltime,
  'canceldatetime': instance.canceldatetime,
  'cancelusercode': instance.cancelusercode,
  'cancelusername': instance.cancelusername,
  'canceldescription': instance.canceldescription,
  'cancelreason': instance.cancelreason,
  'creator_code': instance.creatorcode,
  'creator_name': instance.creatorname,
  'created_at': instance.createdat,
  'modifier_code': instance.modifiercode,
  'modifier_name': instance.modifiername,
  'modified_at': instance.modifiedat,
  'fullvataddress': instance.fullvataddress,
  'fullvatbranchnumber': instance.fullvatbranchnumber,
  'fullvatname': instance.fullvatname,
  'fullvatdocnumber': instance.fullvatdocnumber,
  'fullvattaxid': instance.fullvattaxid,
  'fullvatprint': instance.fullvatprint,
  'isvatregister': instance.isvatregister,
  'printcopybilldatetime': instance.printcopybilldatetime,
  'tablenumber': instance.tablenumber,
  'tableopendatetime': instance.tableopendatetime,
  'tableclosedatetime': instance.tableclosedatetime,
  'mancount': instance.mancount,
  'womancount': instance.womancount,
  'childcount': instance.childcount,
  'tableallacratemode': instance.tableallacratemode,
  'buffetcode': instance.buffetcode,
  'customertelephone': instance.customertelephone,
  'totalqty': instance.totalqty,
  'totaldiscountvatamount': instance.totaldiscountvatamount,
  'totaldiscountexceptvatamount': instance.totaldiscountexceptvatamount,
  'cashiername': instance.cashiername,
  'paycashchange': instance.paycashchange,
  'sumqrcode': instance.sumqrcode,
  'sumcreditcard': instance.sumcreditcard,
  'summoneytransfer': instance.summoneytransfer,
  'sumcheque': instance.sumcheque,
  'sumcoupon': instance.sumcoupon,
  'detaildiscountformula': instance.detaildiscountformula,
  'detailtotalamount': instance.detailtotalamount,
  'detailtotaldiscount': instance.detailtotaldiscount,
  'roundamount': instance.roundamount,
  'totalamountafterdiscount': instance.totalamountafterdiscount,
  'detailtotalamountbeforediscount': instance.detailtotalamountbeforediscount,
  'sumcredit': instance.sumcredit,
  'salechannelcode': instance.salechannelcode,
  'salechannelgp': instance.salechannelgp,
  'salechannelgptype': instance.salechannelgptype,
  'takeaway': instance.takeaway,
  'branch': instance.branch?.toJson(),
  'reftotaloriginal': instance.reftotaloriginal,
  'reftotalcorrect': instance.reftotalcorrect,
  'reftotaldiff': instance.reftotaldiff,
  'slipurl': instance.slipurl,
  'ispos': instance.ispos,
  'isbom': instance.isbom,
  'isdelivery': instance.isdelivery,
  'deliveryamount': instance.deliveryamount,
  'istransport': instance.istransport,
  'transportcode': instance.transportcode,
  'transportamount': instance.transportamount,
  'getpoint': instance.getpoint,
  'usepoint': instance.usepoint,
  'pointdiscountamount': instance.pointdiscountamount,
  'paypointamount': instance.paypointamount,
  'pointscode': instance.pointscode,
  'coupons': instance.coupons?.map((e) => e.toJson()).toList(),
  'totalcouponamount': instance.totalcouponamount,
  'coupondiscountamount': instance.coupondiscountamount,
  'couponcashamount': instance.couponcashamount,
  'docreferences': ?instance.docreferences?.map((e) => e.toJson()).toList(),
  'advancepaymentdocs': instance.advancepaymentdocs
      ?.map((e) => e.toJson())
      .toList(),
  'sumadvancepayment': instance.sumadvancepayment,
  'sumdeposit': instance.sumdeposit,
  'depositdocs': instance.depositdocs?.map((e) => e.toJson()).toList(),
  'purchasetypecode': instance.purchasetypecode,
  'purchasetypenames': instance.purchasetypenames
      ?.map((e) => e.toJson())
      .toList(),
  'currency': instance.currency,
  'currencysymbol': instance.currencysymbol,
  'exchangerate': instance.exchangerate,
  'totalamount_doc': instance.totalamountDoc,
  'totalvalue_doc': instance.totalvalueDoc,
  'totaldiscount_doc': instance.totaldiscountDoc,
  'totalvatvalue_doc': instance.totalvatvalueDoc,
  'totalbeforevat_doc': instance.totalbeforevatDoc,
  'totalaftervat_doc': instance.totalaftervatDoc,
  'doc_currency': instance.docCurrency,
  'doc_currencysymbol': instance.docCurrencySymbol,
  'wht_entries': instance.whtEntries?.map((e) => e.toJson()).toList(),
  'creditdays': instance.creditdays,
  'duedate': instance.duedate,
  'isdelete': instance.isdelete,
};

TransactionDetailModel _$TransactionDetailModelFromJson(
  Map<String, dynamic> json,
) => TransactionDetailModel(
  docdatetime: json['docdatetime'] as String,
  itemguid: json['itemguid'] as String,
  barcode: json['barcode'] as String,
  itemcode: json['itemcode'] as String,
  itemnames: (json['itemnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String,
  qty: (json['qty'] as num).toDouble(),
  price: (json['price'] as num).toDouble(),
  discount: json['discount'] as String,
  sumofcost: (json['sumofcost'] as num).toDouble(),
  remark: json['remark'] as String,
  linenumber: (json['linenumber'] as num).toInt(),
  whcode: json['whcode'] as String,
  shelfcode: json['shelfcode'] as String,
  sumamount: (json['sumamount'] as num).toDouble(),
  locationcode: json['locationcode'] as String,
  totalvaluevat: (json['totalvaluevat'] as num).toDouble(),
  totalqty: (json['totalqty'] as num).toDouble(),
  sumamountexcludevat: (json['sumamountexcludevat'] as num).toDouble(),
  priceexcludevat: (json['priceexcludevat'] as num).toDouble(),
  discountamount: (json['discountamount'] as num).toDouble(),
  standvalue: (json['standvalue'] as num).toDouble(),
  dividevalue: (json['dividevalue'] as num).toDouble(),
  calcflag: (json['calcflag'] as num).toInt(),
  vattype: (json['vattype'] as num).toInt(),
  averagecost: (json['averagecost'] as num).toDouble(),
  ispos: (json['ispos'] as num).toInt(),
  laststatus: (json['laststatus'] as num).toInt(),
  itemtype: (json['itemtype'] as num).toInt(),
  taxtype: (json['taxtype'] as num).toInt(),
  inquirytype: (json['inquirytype'] as num).toInt(),
  multiunit: json['multiunit'] as bool?,
  unitnames: (json['unitnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  whnames: (json['whnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  locationnames: (json['locationnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  towhnames: (json['towhnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  tolocationnames: (json['tolocationnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  towhcode: json['towhcode'] as String?,
  tolocationcode: json['tolocationcode'] as String?,
  docref: json['docref'] as String?,
  docrefdatetime: json['docrefdatetime'] as String?,
  vatcal: (json['vatcal'] as num?)?.toInt(),
  refbarcodes: (json['refbarcodes'] as List<dynamic>?)
      ?.map((e) => BarCodeSubModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  sku: json['sku'] as String?,
  extrajson: json['extrajson'] as String?,
  extrajsonlist: (json['extrajsonlist'] as List<dynamic>?)
      ?.map((e) => ExtraJsonListModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  manufacturerguid: json['manufacturerguid'] as String?,
  description: json['description'] as String?,
  imageuri: json['imageuri'] as String?,
  eventqty: (json['eventqty'] as num?)?.toDouble(),
  priceDoc: (json['price_doc'] as num?)?.toDouble(),
  sumAmountDoc: (json['sumamount_doc'] as num?)?.toDouble(),
  discountamountDoc: (json['discountamount_doc'] as num?)?.toDouble(),
  priceexcludevatDoc: (json['priceexcludevat_doc'] as num?)?.toDouble(),
  sumamountexcludevatDoc: (json['sumamountexcludevat_doc'] as num?)?.toDouble(),
  totalvaluevatDoc: (json['totalvaluevat_doc'] as num?)?.toDouble(),
);

Map<String, dynamic> _$TransactionDetailModelToJson(
  TransactionDetailModel instance,
) => <String, dynamic>{
  'docdatetime': instance.docdatetime,
  'itemguid': instance.itemguid,
  'barcode': instance.barcode,
  'itemcode': instance.itemcode,
  'docref': instance.docref,
  'docrefdatetime': instance.docrefdatetime,
  'itemnames': instance.itemnames?.map((e) => e.toJson()).toList(),
  'unitcode': instance.unitcode,
  'qty': instance.qty,
  'price': instance.price,
  'discount': instance.discount,
  'sumofcost': instance.sumofcost,
  'sumamount': instance.sumamount,
  'remark': instance.remark,
  'linenumber': instance.linenumber,
  'whcode': instance.whcode,
  'whnames': instance.whnames?.map((e) => e.toJson()).toList(),
  'locationcode': instance.locationcode,
  'locationnames': instance.locationnames?.map((e) => e.toJson()).toList(),
  'towhcode': instance.towhcode,
  'towhnames': instance.towhnames?.map((e) => e.toJson()).toList(),
  'tolocationcode': instance.tolocationcode,
  'tolocationnames': instance.tolocationnames?.map((e) => e.toJson()).toList(),
  'shelfcode': instance.shelfcode,
  'totalqty': instance.totalqty,
  'discountamount': instance.discountamount,
  'averagecost': instance.averagecost,
  'totalvaluevat': instance.totalvaluevat,
  'sumamountexcludevat': instance.sumamountexcludevat,
  'priceexcludevat': instance.priceexcludevat,
  'standvalue': instance.standvalue,
  'dividevalue': instance.dividevalue,
  'calcflag': instance.calcflag,
  'vattype': instance.vattype,
  'ispos': instance.ispos,
  'laststatus': instance.laststatus,
  'itemtype': instance.itemtype,
  'taxtype': instance.taxtype,
  'vatcal': instance.vatcal,
  'inquirytype': instance.inquirytype,
  'multiunit': instance.multiunit,
  'unitnames': instance.unitnames?.map((e) => e.toJson()).toList(),
  'refbarcodes': instance.refbarcodes?.map((e) => e.toJson()).toList(),
  'sku': instance.sku,
  'extrajson': instance.extrajson,
  'extrajsonlist': instance.extrajsonlist?.map((e) => e.toJson()).toList(),
  'manufacturerguid': instance.manufacturerguid,
  'description': instance.description,
  'imageuri': instance.imageuri,
  'eventqty': instance.eventqty,
  'price_doc': instance.priceDoc,
  'sumamount_doc': instance.sumAmountDoc,
  'discountamount_doc': instance.discountamountDoc,
  'priceexcludevat_doc': instance.priceexcludevatDoc,
  'sumamountexcludevat_doc': instance.sumamountexcludevatDoc,
  'totalvaluevat_doc': instance.totalvaluevatDoc,
};

MongoTransactionModel _$MongoTransactionModelFromJson(
  Map<String, dynamic> json,
) => MongoTransactionModel(
  collection: json['collection'] as String,
  keyid: json['keyid'] as String,
  body: TransactionModel.fromJson(json['body'] as Map<String, dynamic>),
);

Map<String, dynamic> _$MongoTransactionModelToJson(
  MongoTransactionModel instance,
) => <String, dynamic>{
  'collection': instance.collection,
  'keyid': instance.keyid,
  'body': instance.body.toJson(),
};

TransactionPayModel _$TransactionPayModelFromJson(Map<String, dynamic> json) =>
    TransactionPayModel(
      cashamounttext: json['cashamounttext'] as String?,
      cashamount: (json['cashamount'] as num?)?.toDouble(),
      discountformula: json['discountformula'] as String?,
      discountamount: (json['discountamount'] as num?)?.toDouble(),
      totalafterdiscount: (json['totalafterdiscount'] as num?)?.toDouble(),
      roundamount: (json['roundamount'] as num?)?.toDouble(),
      totalafterround: (json['totalafterround'] as num?)?.toDouble(),
      creditamount: (json['creditamount'] as num?)?.toDouble(),
      paymentcreditcards:
          (json['paymentcreditcards'] as List<dynamic>?)
              ?.map(
                (e) => PayCreditCardModel.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          const [],
      paymenttransfers:
          (json['paymenttransfers'] as List<dynamic>?)
              ?.map((e) => PayTransferModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );

Map<String, dynamic> _$TransactionPayModelToJson(
  TransactionPayModel instance,
) => <String, dynamic>{
  'cashamounttext': instance.cashamounttext,
  'cashamount': instance.cashamount,
  'discountformula': instance.discountformula,
  'discountamount': instance.discountamount,
  'totalafterdiscount': instance.totalafterdiscount,
  'roundamount': instance.roundamount,
  'totalafterround': instance.totalafterround,
  'creditamount': instance.creditamount,
  'paymentcreditcards': instance.paymentcreditcards
      ?.map((e) => e.toJson())
      .toList(),
  'paymenttransfers': instance.paymenttransfers
      ?.map((e) => e.toJson())
      .toList(),
};

PayCouponModel _$PayCouponModelFromJson(Map<String, dynamic> json) =>
    PayCouponModel(
      number: json['number'] as String,
      description: json['description'] as String,
      amount: (json['amount'] as num).toDouble(),
    );

Map<String, dynamic> _$PayCouponModelToJson(PayCouponModel instance) =>
    <String, dynamic>{
      'number': instance.number,
      'description': instance.description,
      'amount': instance.amount,
    };

PayCashModel _$PayCashModelFromJson(Map<String, dynamic> json) => PayCashModel(
  walletid: json['walletid'] as String,
  amount: json['amount'] as String,
);

Map<String, dynamic> _$PayCashModelToJson(PayCashModel instance) =>
    <String, dynamic>{'walletid': instance.walletid, 'amount': instance.amount};

PayCreditCardModel _$PayCreditCardModelFromJson(Map<String, dynamic> json) =>
    PayCreditCardModel(
      amount: (json['amount'] as num?)?.toDouble(),
      cardnumber: json['cardnumber'] as String?,
      chargevalue: (json['chargevalue'] as num?)?.toDouble(),
      chargeword: json['chargeword'] as String?,
      docdatetime: json['docdatetime'] as String?,
      totalnetworth: (json['totalnetworth'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$PayCreditCardModelToJson(PayCreditCardModel instance) =>
    <String, dynamic>{
      'amount': instance.amount,
      'cardnumber': instance.cardnumber,
      'chargevalue': instance.chargevalue,
      'chargeword': instance.chargeword,
      'docdatetime': instance.docdatetime,
      'totalnetworth': instance.totalnetworth,
    };

PayTransferModel _$PayTransferModelFromJson(Map<String, dynamic> json) =>
    PayTransferModel(
      docdatetime: json['docdatetime'] as String?,
      bankcode: json['bankcode'] as String?,
      bookbankcode: json['bookbankcode'] as String?,
      banknames: (json['banknames'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      amount: (json['amount'] as num?)?.toDouble(),
      accountnumber: json['accountnumber'] as String?,
      frombankcode: json['frombankcode'] as String?,
    );

Map<String, dynamic> _$PayTransferModelToJson(PayTransferModel instance) =>
    <String, dynamic>{
      'docdatetime': instance.docdatetime,
      'bankcode': instance.bankcode,
      'bookbankcode': instance.bookbankcode,
      'banknames': instance.banknames?.map((e) => e.toJson()).toList(),
      'accountnumber': instance.accountnumber,
      'amount': instance.amount,
      'frombankcode': instance.frombankcode,
    };

PayChequeModel _$PayChequeModelFromJson(Map<String, dynamic> json) =>
    PayChequeModel(
      duedate: DateTime.parse(json['duedate'] as String),
      bankcode: json['bankcode'] as String,
      bankname: json['bankname'] as String,
      branchnumber: json['branchnumber'] as String,
      chequenumber: json['chequenumber'] as String,
      amount: (json['amount'] as num).toDouble(),
    );

Map<String, dynamic> _$PayChequeModelToJson(PayChequeModel instance) =>
    <String, dynamic>{
      'duedate': instance.duedate.toIso8601String(),
      'bankcode': instance.bankcode,
      'bankname': instance.bankname,
      'branchnumber': instance.branchnumber,
      'chequenumber': instance.chequenumber,
      'amount': instance.amount,
    };

PayDiscountModel _$PayDiscountModelFromJson(Map<String, dynamic> json) =>
    PayDiscountModel(
      code: json['code'] as String,
      description: json['description'] as String,
      formula: json['formula'] as String,
      amount: (json['amount'] as num).toDouble(),
    );

Map<String, dynamic> _$PayDiscountModelToJson(PayDiscountModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'description': instance.description,
      'formula': instance.formula,
      'amount': instance.amount,
    };

PayQrModel _$PayQrModelFromJson(Map<String, dynamic> json) => PayQrModel(
  providercode: json['providercode'] as String? ?? "",
  providername: json['providername'] as String? ?? "",
  description: json['description'] as String? ?? "",
  amount: (json['amount'] as num).toDouble(),
);

Map<String, dynamic> _$PayQrModelToJson(PayQrModel instance) =>
    <String, dynamic>{
      'providercode': instance.providercode,
      'providername': instance.providername,
      'description': instance.description,
      'amount': instance.amount,
    };

TransactionPaidPayModel _$TransactionPaidPayModelFromJson(
  Map<String, dynamic> json,
) => TransactionPaidPayModel(
  shopid: json['shopid'] as String?,
  docno: json['docno'] as String,
  docdatetime: json['docdatetime'] as String,
  doctype: (json['doctype'] as num).toInt(),
  custcode: json['custcode'] as String,
  salecode: json['salecode'] as String,
  salename: json['salename'] as String,
  totalpaymentamount: (json['totalpaymentamount'] as num).toDouble(),
  totalamount: (json['totalamount'] as num).toDouble(),
  totalvalue: (json['totalvalue'] as num).toDouble(),
  totalbalance: (json['totalbalance'] as num).toDouble(),
  details: (json['details'] as List<dynamic>?)
      ?.map(
        (e) =>
            TransactionPaidPayDetailModel.fromJson(e as Map<String, dynamic>),
      )
      .toList(),
  transflag: (json['transflag'] as num).toInt(),
  custnames: (json['custnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  guidfixed: json['guidfixed'] as String?,
  paymentdetailraw: json['paymentdetailraw'] as String?,
  paymentStructs: (json['paymentStructs'] as List<dynamic>?)
      ?.map((e) => BillPayStruct.fromJson(e as Map<String, dynamic>))
      .toList(),
  paycashamount: (json['paycashamount'] as num?)?.toDouble(),
  sumqrcode: (json['sumqrcode'] as num?)?.toDouble(),
  sumcreditcard: (json['sumcreditcard'] as num?)?.toDouble(),
  summoneytransfer: (json['summoneytransfer'] as num?)?.toDouble(),
  sumcheque: (json['sumcheque'] as num?)?.toDouble(),
  sumcoupon: (json['sumcoupon'] as num?)?.toDouble(),
  sumcredit: (json['sumcredit'] as num?)?.toDouble(),
  roundamount: (json['roundamount'] as num?)?.toDouble(),
  paycashchange: (json['paycashchange'] as num?)?.toDouble(),
  branch: json['branch'] == null
      ? null
      : BranchModel.fromJson(json['branch'] as Map<String, dynamic>),
);

Map<String, dynamic> _$TransactionPaidPayModelToJson(
  TransactionPaidPayModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'shopid': instance.shopid,
  'docno': instance.docno,
  'docdatetime': instance.docdatetime,
  'doctype': instance.doctype,
  'transflag': instance.transflag,
  'custcode': instance.custcode,
  'custnames': instance.custnames?.map((e) => e.toJson()).toList(),
  'salecode': instance.salecode,
  'salename': instance.salename,
  'totalpaymentamount': instance.totalpaymentamount,
  'totalamount': instance.totalamount,
  'totalbalance': instance.totalbalance,
  'totalvalue': instance.totalvalue,
  'details': instance.details?.map((e) => e.toJson()).toList(),
  'paymentdetailraw': instance.paymentdetailraw,
  'paymentStructs': instance.paymentStructs?.map((e) => e.toJson()).toList(),
  'paycashamount': instance.paycashamount,
  'sumqrcode': instance.sumqrcode,
  'sumcreditcard': instance.sumcreditcard,
  'summoneytransfer': instance.summoneytransfer,
  'sumcheque': instance.sumcheque,
  'sumcoupon': instance.sumcoupon,
  'sumcredit': instance.sumcredit,
  'roundamount': instance.roundamount,
  'paycashchange': instance.paycashchange,
  'branch': instance.branch?.toJson(),
};

TransactionPaidPayDetailModel _$TransactionPaidPayDetailModelFromJson(
  Map<String, dynamic> json,
) => TransactionPaidPayDetailModel(
  selected: json['selected'] as bool? ?? false,
  docno: json['docno'] as String,
  docdatetime: json['docdatetime'] as String,
  transflag: (json['transflag'] as num).toInt(),
  value: (json['value'] as num).toDouble(),
  balance: (json['balance'] as num).toDouble(),
  paymentamount: (json['paymentamount'] as num).toDouble(),
);

Map<String, dynamic> _$TransactionPaidPayDetailModelToJson(
  TransactionPaidPayDetailModel instance,
) => <String, dynamic>{
  'selected': instance.selected,
  'docno': instance.docno,
  'docdatetime': instance.docdatetime,
  'transflag': instance.transflag,
  'value': instance.value,
  'balance': instance.balance,
  'paymentamount': instance.paymentamount,
};

BillPayStruct _$BillPayStructFromJson(Map<String, dynamic> json) =>
    BillPayStruct(
      doc_mode: (json['doc_mode'] as num?)?.toInt(),
      trans_flag: (json['trans_flag'] as num?)?.toInt(),
      bank_code: json['bank_code'] as String?,
      bank_name: json['bank_name'] as String?,
      book_bank_code: json['book_bank_code'] as String?,
      card_number: json['card_number'] as String?,
      approved_code: json['approved_code'] as String?,
      doc_date_time: json['doc_date_time'] == null
          ? null
          : DateTime.parse(json['doc_date_time'] as String),
      branch_number: json['branch_number'] as String?,
      bank_reference: json['bank_reference'] as String?,
      due_date: json['due_date'] == null
          ? null
          : DateTime.parse(json['due_date'] as String),
      cheque_number: json['cheque_number'] as String?,
      code: json['code'] as String?,
      description: json['description'] as String?,
      number: json['number'] as String?,
      reference_one: json['reference_one'] as String?,
      reference_two: json['reference_two'] as String?,
      provider_code: json['provider_code'] as String?,
      provider_name: json['provider_name'] as String?,
      amount: (json['amount'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$BillPayStructToJson(BillPayStruct instance) =>
    <String, dynamic>{
      'doc_mode': instance.doc_mode,
      'trans_flag': instance.trans_flag,
      'bank_code': instance.bank_code,
      'bank_name': instance.bank_name,
      'book_bank_code': instance.book_bank_code,
      'card_number': instance.card_number,
      'approved_code': instance.approved_code,
      'doc_date_time': instance.doc_date_time?.toIso8601String(),
      'branch_number': instance.branch_number,
      'bank_reference': instance.bank_reference,
      'due_date': instance.due_date?.toIso8601String(),
      'cheque_number': instance.cheque_number,
      'code': instance.code,
      'description': instance.description,
      'number': instance.number,
      'reference_one': instance.reference_one,
      'reference_two': instance.reference_two,
      'provider_code': instance.provider_code,
      'provider_name': instance.provider_name,
      'amount': instance.amount,
    };

ExtraJsonListModel _$ExtraJsonListModelFromJson(Map<String, dynamic> json) =>
    ExtraJsonListModel(
      barcode: json['barcode'] as String?,
      item_code: json['item_code'] as String?,
      unit_code: json['unit_code'] as String?,
      unit_name: json['unit_name'] as String?,
      qty: (json['qty'] as num?)?.toDouble(),
      price: (json['price'] as num?)?.toDouble(),
      total_amount: (json['total_amount'] as num?)?.toDouble(),
      is_except_vat: json['is_except_vat'] as bool?,
      vat_type: (json['vat_type'] as num?)?.toInt(),
      price_exclude_vat: (json['price_exclude_vat'] as num?)?.toDouble(),
      item_name: json['item_name'] as String?,
      itemnames: (json['itemnames'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ExtraJsonListModelToJson(ExtraJsonListModel instance) =>
    <String, dynamic>{
      'barcode': instance.barcode,
      'item_code': instance.item_code,
      'unit_code': instance.unit_code,
      'unit_name': instance.unit_name,
      'qty': instance.qty,
      'price': instance.price,
      'total_amount': instance.total_amount,
      'is_except_vat': instance.is_except_vat,
      'vat_type': instance.vat_type,
      'price_exclude_vat': instance.price_exclude_vat,
      'item_name': instance.item_name,
      'itemnames': instance.itemnames,
    };

GetCustcodeTransationModel _$GetCustcodeTransationModelFromJson(
  Map<String, dynamic> json,
) => GetCustcodeTransationModel(
  shopid: json['shopid'] as String?,
  parid: json['parid'] as String?,
  docno: json['docno'] as String?,
  docdate: json['docdate'] == null
      ? null
      : DateTime.parse(json['docdate'] as String),
  creditorcode: json['creditorcode'] as String?,
  inquirytype: json['inquirytype'] as String?,
  transflag8: (json['transflag8'] as num?)?.toInt(),
  totalvalue: json['totalvalue'] as String?,
  totalbeforevat: json['totalbeforevat'] as String?,
  totalvatvalue: json['totalvatvalue'] as String?,
  totalexceptvat: json['totalexceptvat'] as String?,
  totalaftervat: json['totalaftervat'] as String?,
  totalamount: json['totalamount'] as String?,
  paidamount: json['paidamount'] as String?,
  balanceamount: json['balanceamount'] as String?,
  status: (json['status'] as num?)?.toInt(),
  iscancel: json['iscancel'] as bool?,
  guidfixed: json['guidfixed'] as String?,
);

Map<String, dynamic> _$GetCustcodeTransationModelToJson(
  GetCustcodeTransationModel instance,
) => <String, dynamic>{
  'shopid': instance.shopid,
  'parid': instance.parid,
  'docno': instance.docno,
  'docdate': instance.docdate?.toIso8601String(),
  'creditorcode': instance.creditorcode,
  'inquirytype': instance.inquirytype,
  'transflag8': instance.transflag8,
  'totalvalue': instance.totalvalue,
  'totalbeforevat': instance.totalbeforevat,
  'totalvatvalue': instance.totalvatvalue,
  'totalexceptvat': instance.totalexceptvat,
  'totalaftervat': instance.totalaftervat,
  'totalamount': instance.totalamount,
  'paidamount': instance.paidamount,
  'balanceamount': instance.balanceamount,
  'status': instance.status,
  'iscancel': instance.iscancel,
  'guidfixed': instance.guidfixed,
};

CouponModel _$CouponModelFromJson(Map<String, dynamic> json) => CouponModel(
  couponno: json['couponno'] as String?,
  couponamount: (json['couponamount'] as num?)?.toDouble(),
  coupondescription: json['coupondescription'] as String?,
  coupontype: json['coupontype'] as String?,
);

Map<String, dynamic> _$CouponModelToJson(CouponModel instance) =>
    <String, dynamic>{
      'couponno': instance.couponno,
      'couponamount': instance.couponamount,
      'coupondescription': instance.coupondescription,
      'coupontype': instance.coupontype,
    };

DocReferenceModel _$DocReferenceModelFromJson(Map<String, dynamic> json) =>
    DocReferenceModel(
      docdatetime: json['docdatetime'] as String,
      docno: json['docno'] as String,
      guidfixed: json['guidfixed'] as String,
    );

Map<String, dynamic> _$DocReferenceModelToJson(DocReferenceModel instance) =>
    <String, dynamic>{
      'docdatetime': instance.docdatetime,
      'docno': instance.docno,
      'guidfixed': instance.guidfixed,
    };

AdvancePaymentDocModel _$AdvancePaymentDocModelFromJson(
  Map<String, dynamic> json,
) => AdvancePaymentDocModel(
  balance: (json['balance'] as num).toDouble(),
  docdatetime: json['docdatetime'] as String,
  docno: json['docno'] as String,
  guidfixed: json['guidfixed'] as String,
  remark: json['remark'] as String,
  totalamount: (json['totalamount'] as num).toDouble(),
  useamount: (json['useamount'] as num).toDouble(),
);

Map<String, dynamic> _$AdvancePaymentDocModelToJson(
  AdvancePaymentDocModel instance,
) => <String, dynamic>{
  'balance': instance.balance,
  'docdatetime': instance.docdatetime,
  'docno': instance.docno,
  'guidfixed': instance.guidfixed,
  'remark': instance.remark,
  'totalamount': instance.totalamount,
  'useamount': instance.useamount,
};

DepositDocModel _$DepositDocModelFromJson(Map<String, dynamic> json) =>
    DepositDocModel(
      balance: (json['balance'] as num).toDouble(),
      docdatetime: json['docdatetime'] as String,
      docno: json['docno'] as String,
      guidfixed: json['guidfixed'] as String,
      remark: json['remark'] as String,
      totalamount: (json['totalamount'] as num).toDouble(),
      useamount: (json['useamount'] as num).toDouble(),
    );

Map<String, dynamic> _$DepositDocModelToJson(DepositDocModel instance) =>
    <String, dynamic>{
      'balance': instance.balance,
      'docdatetime': instance.docdatetime,
      'docno': instance.docno,
      'guidfixed': instance.guidfixed,
      'remark': instance.remark,
      'totalamount': instance.totalamount,
      'useamount': instance.useamount,
    };
