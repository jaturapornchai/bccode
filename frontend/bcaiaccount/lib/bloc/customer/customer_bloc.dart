import 'dart:convert';

import 'package:smlaicloud/model/customer_request_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/customer_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/customer_repository.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'dart:io';

part 'customer_event.dart';
part 'customer_state.dart';

class CustomerBloc extends Bloc<CustomerEvent, CustomerState> {
  final CustomerRepository _customerRepository;

  CustomerBloc({required CustomerRepository customerRepository})
      : _customerRepository = customerRepository,
        super(CustomerInitial()) {
    on<CustomerLoadList>(onCustomerLoad);
    on<SupplierLoadList>(onSupplierLoad);
    on<CustomerSave>(onCustomerSave);
    on<CustomerWithImageSave>(onCustomerWithImageSave);
    on<CustomerUpdate>(onCustomerUpdate);
    on<CustomerWithImageUpdate>(onCustomerWithImageUpdate);
    on<CustomerDelete>(customerDelete);
    on<CustomerDeleteMany>(customerDeleteMany);
    on<CustomerGet>(onCustomerGet);
  }

  void onCustomerLoad(
      CustomerLoadList event, Emitter<CustomerState> emit) async {
    emit(CustomerInProgress());

    try {
      final results = await _customerRepository.getCustomerList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<CustomerModel> customers = (results.data as List)
            .map((customer) => CustomerModel.fromJson(customer))
            .toList();
        emit(CustomerLoadSuccess(customers: customers));
      } else {
        emit(const CustomerLoadFailed(message: 'Customer Group Not Found'));
      }
    } catch (e) {
      emit(CustomerLoadFailed(message: e.toString()));
    }
  }

  void onSupplierLoad(
      SupplierLoadList event, Emitter<CustomerState> emit) async {
    emit(CustomerInProgress());

    try {
      final results = await _customerRepository.getSupplierList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<CustomerModel> customers = (results.data as List)
            .map((customer) => CustomerModel.fromJson(customer))
            .toList();
        emit(CustomerLoadSuccess(customers: customers));
      } else {
        emit(const CustomerLoadFailed(message: 'Customer Group Not Found'));
      }
    } catch (e) {
      emit(CustomerLoadFailed(message: e.toString()));
    }
  }

  void customerDelete(CustomerDelete event, Emitter<CustomerState> emit) async {
    emit(CustomerDeleteInProgress());
    try {
      await _customerRepository.deleteCustomer(event.guid);

      emit(CustomerDeleteSuccess());
    } catch (e) {
      // emit(CustomerDeleteFailure(message: e.toString()));
    }
  }

  void customerDeleteMany(
      CustomerDeleteMany event, Emitter<CustomerState> emit) async {
    emit(CustomerDeleteManyInProgress());
    try {
      await _customerRepository.deleteCustomerMany(event.guid);

      emit(CustomerDeleteManySuccess());
    } catch (e) {
      // emit(CustomerDeleteFailure(message: e.toString()));
    }
  }

  void onCustomerSave(CustomerSave event, Emitter<CustomerState> emit) async {
    emit(CustomerSaveInProgress());
    try {
      CustomerModel customer = event.customer;

      CustomerRequestModel customerRequestModel = CustomerRequestModel(
          guidfixed: customer.guidfixed,
          code: customer.code,
          names: customer.names,
          customertype: customer.customertype,
          branchnumber: customer.branchnumber,
          personaltype: customer.customertype,
          addressforbilling: customer.addressforbilling,
          addressforshipping: customer.addressforshipping,
          images: customer.images,
          taxid: customer.taxid,
          email: customer.email,
          iscreditor: customer.iscreditor,
          isdebtor: customer.isdebtor,
          creditday: customer.creditday,
          fundcode: customer.fundcode);

      for (var element in customer.groups) {
        customerRequestModel.groups.add(element.guidfixed);
      }

      await _customerRepository.saveCustomer(customerRequestModel);
      emit(CustomerSaveSuccess());
    } catch (e) {
      emit(CustomerSaveFailed(message: e.toString()));
    }
  }

  void onCustomerWithImageSave(
      CustomerWithImageSave event, Emitter<CustomerState> emit) async {
    emit(CustomerSaveInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imageFile.isNotEmpty) {
        for (int i = 0; i < event.imageFile.length; i++) {
          if (event.imageFile[i].uri.toString() != '') {
            ApiResponse result = await _customerRepository.uploadImage(
                event.imageFile[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CustomerSaveFailed(message: result.message));
            }
          }
        }

        if (images.length == event.imageFile.length) {
          // CustomerModel customerModel = event.customer;
          // customerModel.images = images;
          // customerModel.addressForShipping.guid = "";
          // for (int i = 0; i < customerModel.addressForShipping.length; i++) {
          //   customerModel.addressForShipping[i].guid = "";
          // }

          // print(customerModel);
          // await _customerRepository.saveCustomer(customerModel);
          // emit(CustomerSaveSuccess());
        } else {
          emit(const CustomerSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CustomerSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CustomerSaveFailed(message: error['message']));
    }
  }

  void onCustomerUpdate(
      CustomerUpdate event, Emitter<CustomerState> emit) async {
    emit(CustomerUpdateInProgress());
    try {
      CustomerModel customer = event.customerModel;

      CustomerRequestModel customerRequestModel = CustomerRequestModel(
          guidfixed: customer.guidfixed,
          code: customer.code,
          names: customer.names,
          customertype: customer.customertype,
          branchnumber: customer.branchnumber,
          personaltype: customer.customertype,
          addressforbilling: customer.addressforbilling,
          addressforshipping: customer.addressforshipping,
          images: customer.images,
          taxid: customer.taxid,
          email: customer.email,
          iscreditor: customer.iscreditor,
          isdebtor: customer.isdebtor,
          creditday: customer.creditday,
          fundcode: customer.fundcode);

      for (var element in customer.groups) {
        customerRequestModel.groups.add(element.guidfixed);
      }

      await _customerRepository.updateCustomer(
          event.guid, customerRequestModel);
      emit(CustomerUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CustomerUpdateFailed(message: error['message']));
    }
  }

  void onCustomerWithImageUpdate(
      CustomerWithImageUpdate event, Emitter<CustomerState> emit) async {
    emit(CustomerUpdateInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imagesUris.isNotEmpty) {
        for (int i = 0; i < event.imagesUris.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _customerRepository.uploadImage(
                event.imageFiles[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CustomerUpdateFailed(message: result.message));
            }
          } else if (event.imagesUris[i].uri != '') {
            images.add(ImagesModel(uri: event.imagesUris[i].uri, xorder: i));
          }
        }

        if (images.isNotEmpty) {
          // CustomerModel customerModel = event.customer;
          // customerModel.images = images;
          // customerModel.addressForBilling.guid = "";
          // for (int i = 0; i < customerModel.addressForBilling.length; i++) {
          //   customerModel.addressForBilling[i].guid = "";
          // }

          // print(customerModel);
          // await _customerRepository.updateCustomer(event.guid, event.customer);
          // emit(CustomerUpdateSuccess());
        } else {
          emit(const CustomerUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CustomerUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(CustomerUpdateFailed(message: e.toString()));
    }
  }

  void onCustomerGet(CustomerGet event, Emitter<CustomerState> emit) async {
    emit(CustomerGetInProgress());
    try {
      final result = await _customerRepository.getCustomer(event.guid);
      if (result.success) {
        CustomerModel customer = CustomerModel.fromJson(result.data);
        emit(CustomerGetSuccess(customer: customer));
      } else {
        emit(const CustomerGetFailed(message: 'Customer Not Found'));
      }
    } catch (e) {
      // emit(CustomerDeleteFailure(message: e.toString()));
    }
  }
}
