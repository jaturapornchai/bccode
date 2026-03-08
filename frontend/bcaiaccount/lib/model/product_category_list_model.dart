import 'package:smlaicloud/model/product_category_model.dart';

class ProductCategoryListModel {
  late ProductCategoryModel detail;
  late List<ProductCategoryListModel> childCategories;
  late bool isExpand = true;
  late bool isMoveUp = false;
  late bool isMoveDown = false;

  ProductCategoryListModel({
    required this.detail,
    required this.childCategories,
  });
}
