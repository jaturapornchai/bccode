# UI Components & Conventions

## Logger

Always use `AppLogger` — never use `print()`.

```dart
import 'package:smlaicloud/utils/logger/app_logger.dart';

AppLogger.info('message');
AppLogger.debug('message: $value');
AppLogger.error('error: $e');
AppLogger.warning('warning');
```

## Loading / Error / Empty States Pattern

Every page fetching API data must handle all 3 states:

```dart
BlocBuilder<FeatureBloc, FeatureState>(
  builder: (context, state) {
    // Loading
    if (state is FeatureInProgress) {
      return const Center(child: CircularProgressIndicator());
    }
    // Error
    if (state is FeatureLoadFailed) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.red),
            const SizedBox(height: 8),
            Text(state.message),
            TextButton(
              onPressed: () => context.read<FeatureBloc>().add(
                FeatureLoadList(shopId: shopId)),
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }
    // Empty
    if (state is FeatureLoadSuccess && state.items.isEmpty) {
      return const Center(child: Text('No data'));
    }
    // Data
    if (state is FeatureLoadSuccess) {
      return FeatureListWidget(items: state.items);
    }
    return const SizedBox.shrink();
  },
)
```

## Form Save/Update Pattern

Use `BlocConsumer` when widget needs both rebuild and side-effect listening:

```dart
BlocConsumer<FeatureBloc, FeatureState>(
  listener: (context, state) {
    if (state is FeatureSaveSuccess || state is FeatureUpdateSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Saved successfully')),
      );
      Navigator.pop(context, true); // Return true for caller to refresh
    }
    if (state is FeatureSaveFailed || state is FeatureUpdateFailed) {
      final msg = state is FeatureSaveFailed
          ? (state as FeatureSaveFailed).message
          : (state as FeatureUpdateFailed).message;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg), backgroundColor: Colors.red),
      );
    }
  },
  builder: (context, state) {
    final isLoading = state is FeatureSaveInProgress || state is FeatureUpdateInProgress;
    return ElevatedButton(
      onPressed: isLoading ? null : _onSubmit,
      child: isLoading
          ? const SizedBox(width: 20, height: 20,
              child: CircularProgressIndicator(strokeWidth: 2))
          : const Text('Save'),
    );
  },
)
```

## Document Status Badge

```dart
Widget buildStatusBadge(String status) {
  final config = {
    'draft':     ('Draft',     Colors.grey),
    'pending':   ('Pending',   Colors.orange),
    'approved':  ('Approved',  Colors.green),
    'completed': ('Completed', Colors.blue),
    'rejected':  ('Rejected',  Colors.red),
    'cancelled': ('Cancelled', Colors.red.shade300),
  };
  final (label, color) = config[status] ?? (status, Colors.grey);
  return Container(
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
    decoration: BoxDecoration(
      color: color.withOpacity(0.1),
      borderRadius: BorderRadius.circular(4),
      border: Border.all(color: color),
    ),
    child: Text(label, style: TextStyle(color: color, fontSize: 12)),
  );
}
```

## Action Button Visibility (by Document State)

```dart
// Always check state before showing buttons
final canEdit   = doc.docstatus == 'draft' || doc.docstatus == 'rejected';
final canSubmit = doc.docstatus == 'draft';
final canApprove = doc.docstatus == 'pending' && currentUser.isApprover;
final canCancel = !['completed', 'cancelled'].contains(doc.docstatus);

if (canEdit) ElevatedButton(onPressed: _onEdit, child: Text('Edit')),
if (canSubmit) ElevatedButton(onPressed: _onSubmit, child: Text('Submit')),
if (canApprove) ElevatedButton(onPressed: _onApprove, child: Text('Approve')),
```

## Navigation

```dart
// Navigate to new screen
Navigator.push(context, MaterialPageRoute(builder: (_) => FeatureScreen()));

// Return with refresh
final result = await Navigator.push(...);
if (result == true) {
  context.read<FeatureBloc>().add(FeatureLoadList(shopId: shopId));
}

// Named routes (if project uses app_routes.dart)
Navigator.pushNamed(context, AppRoutes.featureDetail, arguments: {'id': guid});
```

## Build Commands

```bash
# Run DEV
flutter run -t lib/main_bcaidev.dart

# Build Web DEV
flutter build web -t lib/main_bcaidev.dart --release

# Build Web PROD
flutter build web -t lib/main_bcaiprod.dart --release

# Build Windows DEV
flutter build windows -t lib/main_bcaidev.dart --release
```

## Code Conventions

- **Language:** Comments and logs can be Thai, but variable/class/function names must be English
- **User-facing text:** Thai always (error messages, SnackBar, Dialog)
- **Forbidden:** Hardcode URLs, API keys, shopid
- **Forbidden:** Flutter calling AI API providers directly (Groq, OpenAI, Gemini)

## Transaction Detail Table — Reorder Pattern

All transaction edit screens (purchase, sales, transfer, stock adjustment, etc.) use `DocumentProductListWidget` for the line items table.

### Reorder (move items up/down)

**Header column:** Add `reorder` header after `delete` in every headers list:

```dart
headers = [
  global.DataTableHeader(code: "delete", label: "", width: 5, ...),
  global.DataTableHeader(code: "reorder", label: "", width: 4, textAlign: TextAlign.center, alignment: Alignment.center),
  global.DataTableHeader(code: "line_number", label: ..., width: 10, ...),
  // ... other columns
];
```

**Callback:** Pass `onReorderItem` to `DocumentProductListWidget`:

```dart
DocumentProductListWidget(
  // ...
  onReorderItem: (int oldIndex, int newIndex) {
    setState(() {
      final item = screenData.details!.removeAt(oldIndex);
      screenData.details!.insert(newIndex, item);
      for (int i = 0; i < screenData.details!.length; i++) {
        screenData.details![i].linenumber = i + 1;
      }
    });
  },
  // ...
);
```

**Column width mapping:** In `Table.columnWidths` add `reorder`:

```dart
columnWidths: {
  for (int i = 0; i < headers.length; i++)
    i: (headers[i].code == 'line_number')
        ? const FixedColumnWidth(50.0)
        : (headers[i].code == 'reorder')
            ? const FixedColumnWidth(30.0)
            : FlexColumnWidth(headers[i].width)
},
```

**Files to modify when adding reorder:**
- `table_header_widget.dart` — columnWidths
- `document_product_list_widget.dart` — columnWidths + cell render (already exists)
- Every `*_edit_screen.dart` — headers list + onReorderItem callback

**Screens with reorder already implemented:**
- `transaction_edit.dart` — Sales/Purchase/Transfer/Stock Adjustment
- `purchaseorder_edit_screen.dart` — Purchase Order
- `purchaserequisition_edit_screen.dart` — Purchase Requisition
- `rfq_edit_screen.dart` — Request for Quotation
- `quotation_edit_screen.dart` — Quotation
