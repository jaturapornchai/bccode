part of 'knowledge_base_cubit.dart';

class KnowledgeBaseState {
  final List<DocumentModel> documents;
  final bool isLoading;
  final String? errorMessage;

  const KnowledgeBaseState({
    this.documents = const [],
    this.isLoading = false,
    this.errorMessage,
  });

  KnowledgeBaseState copyWith({
    List<DocumentModel>? documents,
    bool? isLoading,
    String? errorMessage,
  }) {
    return KnowledgeBaseState(
      documents: documents ?? this.documents,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
    );
  }
}
