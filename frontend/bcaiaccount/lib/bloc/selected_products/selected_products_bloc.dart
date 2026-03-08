import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'package:smlaicloud/screens/components/selected_product_grid_item.dart';

part 'selected_products_event.dart';
part 'selected_products_state.dart';

class SelectedProductsBloc extends Bloc<SelectedProductsEvent, SelectedProductsState> {
  SelectedProductsBloc() : super(const SelectedProductsInitial()) {
    on<AddProductToSelection>(_onAddProductToSelection);
    on<AddProductsFromShelf>(_onAddProductsFromShelf);
    on<AddAllProducts>(_onAddAllProducts);
    on<RemoveProductFromSelection>(_onRemoveProductFromSelection);
    on<UpdateProductCopies>(_onUpdateProductCopies);
    on<UpdateAllProductsCopies>(_onUpdateAllProductsCopies);
    on<ClearAllProducts>(_onClearAllProducts);
  }

  void _onAddProductToSelection(
    AddProductToSelection event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    if (currentState is SelectedProductsLoaded) {
      if (!currentState.isProductSelected(event.product.barcode!)) {
        final updatedProducts = List<ProductWithCopies>.from(currentState.selectedProducts)
          ..add(ProductWithCopies(
            product: SearchCodeAndNameAndUnitModel(
              barcode: event.product.barcode!,
              code: event.product.itemcode ?? event.product.barcode!, // ใช้ barcode ถ้า itemcode เป็น null
              name: event.product.names!,
              unitcode: event.product.itemunitcode,
              unitname: event.product.itemunitnames ?? [],
              shelfCode: event.product.shelfCode, // เพิ่มข้อมูลชั้นวาง
              shelfName: event.product.shelfName, // เพิ่มข้อมูลชั้นวาง
            ),
            copies: 1,
          ));
        emit(SelectedProductsLoaded(selectedProducts: updatedProducts));
      }
    } else {
      // Initial state
      emit(SelectedProductsLoaded(
        selectedProducts: [
          ProductWithCopies(
            product: SearchCodeAndNameAndUnitModel(
              barcode: event.product.barcode!,
              code: event.product.itemcode ?? event.product.barcode!, // ใช้ barcode ถ้า itemcode เป็น null
              name: event.product.names!,
              unitcode: event.product.itemunitcode,
              unitname: event.product.itemunitnames ?? [],
              shelfCode: event.product.shelfCode, // เพิ่มข้อมูลชั้นวาง
              shelfName: event.product.shelfName, // เพิ่มข้อมูลชั้นวาง
            ),
            copies: 1,
          ),
        ],
      ));
    }
  }

  void _onAddProductsFromShelf(
    AddProductsFromShelf event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    final currentProducts = currentState is SelectedProductsLoaded ? List<ProductWithCopies>.from(currentState.selectedProducts) : <ProductWithCopies>[];

    for (var shelfProduct in event.products) {
      final isAlreadySelected = currentProducts.any(
        (item) => item.product.barcode == shelfProduct.barcode,
      );

      if (!isAlreadySelected) {
        currentProducts.add(ProductWithCopies(
          product: SearchCodeAndNameAndUnitModel(
            barcode: shelfProduct.barcode,
            code: shelfProduct.barcode,
            name: shelfProduct.names,
            unitcode: shelfProduct.unitcode ?? '',
            unitname: shelfProduct.unitnames ?? [],
            shelfCode: event.shelfCode,
            shelfName: event.shelfName,
          ),
          copies: 1,
        ));
      }
    }

    emit(SelectedProductsLoaded(selectedProducts: currentProducts));
  }

  void _onAddAllProducts(
    AddAllProducts event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    final currentProducts = currentState is SelectedProductsLoaded ? List<ProductWithCopies>.from(currentState.selectedProducts) : <ProductWithCopies>[];

    for (var product in event.products) {
      final isAlreadySelected = currentProducts.any(
        (item) => item.product.barcode == product.barcode,
      );

      if (!isAlreadySelected) {
        currentProducts.add(ProductWithCopies(
          product: SearchCodeAndNameAndUnitModel(
            barcode: product.barcode!,
            code: product.itemcode ?? product.barcode!, // ใช้ barcode ถ้า itemcode เป็น null
            name: product.names!,
            unitcode: product.itemunitcode,
            unitname: product.itemunitnames ?? [],
            shelfCode: product.shelfCode, // เพิ่มข้อมูลชั้นวาง
            shelfName: product.shelfName, // เพิ่มข้อมูลชั้นวาง
          ),
          copies: 1,
        ));
      }
    }

    emit(SelectedProductsLoaded(selectedProducts: currentProducts));
  }

  void _onRemoveProductFromSelection(
    RemoveProductFromSelection event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    if (currentState is SelectedProductsLoaded) {
      final updatedProducts = currentState.selectedProducts.where((item) => item.product.barcode != event.barcode).toList();
      emit(SelectedProductsLoaded(selectedProducts: updatedProducts));
    }
  }

  void _onUpdateProductCopies(
    UpdateProductCopies event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    if (currentState is SelectedProductsLoaded) {
      final updatedProducts = currentState.selectedProducts.map((item) {
        if (item.product.barcode == event.barcode) {
          item.copies = event.copies;
        }
        return item;
      }).toList();
      emit(SelectedProductsLoaded(selectedProducts: updatedProducts));
    }
  }

  void _onUpdateAllProductsCopies(
    UpdateAllProductsCopies event,
    Emitter<SelectedProductsState> emit,
  ) {
    final currentState = state;
    if (currentState is SelectedProductsLoaded) {
      final updatedProducts = currentState.selectedProducts.map((item) {
        item.copies = event.copies;
        return item;
      }).toList();
      emit(SelectedProductsLoaded(selectedProducts: updatedProducts));
    }
  }

  void _onClearAllProducts(
    ClearAllProducts event,
    Emitter<SelectedProductsState> emit,
  ) {
    emit(const SelectedProductsLoaded(selectedProducts: []));
  }
}
