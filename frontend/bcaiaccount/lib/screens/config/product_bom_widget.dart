// preview_screen_widget.dart
import 'package:flutter/material.dart';
import 'package:graphview/GraphView.dart';
import 'package:smlaicloud/model/product_bom_model.dart';
import 'package:smlaicloud/global.dart' as global;
// Import any other dependencies that your widget needs.

class ProductBomWidget extends StatelessWidget {
  final ProductBomModel productBom;

  const ProductBomWidget({super.key, required this.productBom});

  @override
  Widget build(BuildContext context) {
    Size size = MediaQuery.of(context).size;
    var graph = buildGraph(productBom);
    BuchheimWalkerConfiguration builder = BuchheimWalkerConfiguration()
      ..siblingSeparation = (100)
      ..levelSeparation = (100)
      ..subtreeSeparation = (100)
      ..orientation = BuchheimWalkerConfiguration.ORIENTATION_TOP_BOTTOM;

    return SingleChildScrollView(
      child: SizedBox(
        height: size.height,
        width: size.width,
        child: (productBom.guidfixed!.isEmpty)
            ? Center(
                child: Text(global.language('no_bom_this_product'),
                    style: TextStyle(color: global.theme.negativeHighlightTextColor, fontSize: 20)))
            : InteractiveViewer(
                constrained: false,
                boundaryMargin: EdgeInsets.all(100),
                minScale: 1.0, // Set minScale to 1.0 to disable zooming
                maxScale: 1.0, // Set maxScale to 1.0 to disable zooming
                child: GraphView(
                  graph: graph,
                  algorithm: BuchheimWalkerAlgorithm(
                      builder, TreeEdgeRenderer(builder)),
                  paint: Paint()
                    ..color = global.theme.infoHighlightTextColor
                    ..strokeWidth = 2
                    ..style = PaintingStyle.stroke,
                  builder: (Node node) {
                    var guid = node.key!.value as String;
                    var matchingProductBom =
                        findProductBomByGuid(productBom, guid);
                    return rectangleWidget(matchingProductBom ?? productBom);
                  },
                ),
              ),
      ),
    );
  }

  Graph buildGraph(ProductBomModel rootBom) {
    var graph = Graph()..isTree = true;
    _createNodesAndEdges(graph, rootBom, null);
    return graph;
  }

  void _createNodesAndEdges(
      Graph graph, ProductBomModel bom, Node? parentNode) {
    var node = Node.Id(bom.guidfixed);
    graph.addNode(node);
    if (parentNode != null) {
      graph.addEdge(parentNode, node);
    }

    for (var childBom in bom.bom ?? []) {
      _createNodesAndEdges(graph, childBom, node);
    }
  }

  ProductBomModel? findProductBomByGuid(ProductBomModel bom, String guid) {
    if (bom.guidfixed == guid) {
      return bom;
    }
    for (var child in bom.bom ?? []) {
      var result = findProductBomByGuid(child, guid);
      if (result != null) return result;
    }
    return null;
  }

  Widget rectangleWidget(ProductBomModel productBom) {
    return Container(
      padding: EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: global.theme.columnHeaderTextColor,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Image.network(
            productBom.imageuri!,
            fit: BoxFit.fitHeight,
            height: 50,
            errorBuilder: (context, error, stackTrace) => Icon(
                Icons.image_not_supported,
                size: 50,
                color: global.theme.onPrimaryColor),
          ),
          Text(' ${productBom.barcode}',
              style: TextStyle(color: global.theme.onPrimaryColor)),
          Text(global.activeLangName(productBom.names!),
              style: TextStyle(color: global.theme.onPrimaryColor)),
          Text(
              '${global.language('qty')} ${productBom.qty} ${global.activeLangName(productBom.itemunitnames!)}',
              style: TextStyle(color: global.theme.onPrimaryColor)),
        ],
      ),
    );
  }
}
