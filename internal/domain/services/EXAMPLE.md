# Domain Services — Pseudocode Examples

Two domain services: a pricing calculator and an inventory allocator.
Both are stateless and accept all inputs as parameters.

---

## 1. PricingService

Calculates the final price for a cart given applicable discounts.
Logic spans `Cart`, `Product`, and `Discount` — no single entity owns it.

```
DomainService PricingService {

  // Calculate total with all applicable discounts applied
  calculateTotal(items: List<CartItem>, discounts: List<Discount>) → Money:

    subtotal = Money.zero(items[0].price.currency)

    for item in items:
      lineTotal = item.price.multiply(item.quantity)
      subtotal  = subtotal.add(lineTotal)

    discountAmount = self.applyDiscounts(subtotal, discounts)
    total = subtotal.subtract(discountAmount)

    return total

  // Internal: resolve discount amount from all applicable rules
  private applyDiscounts(subtotal: Money, discounts: List<Discount>) → Money:
    totalDiscount = Money.zero(subtotal.currency)

    for discount in discounts:
      if discount.type == PERCENTAGE:
        amount = subtotal.multiplyByPercent(discount.value)
      else if discount.type == FIXED:
        amount = discount.fixedAmount
      else:
        raise DomainError("Unknown discount type: " + discount.type)

      totalDiscount = totalDiscount.add(amount)

    // Discount cannot exceed subtotal
    if totalDiscount.isGreaterThan(subtotal):
      return subtotal

    return totalDiscount
}
```

**What it does NOT do:**
- Does not load products from a database
- Does not call a discount API
- Does not check inventory
- Does not know about HTTP, gRPC, or any transport

---

## 2. InventoryAllocator

Determines which warehouse should fulfill an order, based on stock levels and
shipping proximity. Logic spans `Order`, `Warehouse`, and `Stock` — belongs to
no single entity.

```
DomainService InventoryAllocator {

  // Find the best warehouse to fulfill all items in the order
  allocate(order: Order, warehouses: List<Warehouse>) → AllocationPlan:

    eligible = []

    for warehouse in warehouses:
      if self.canFulfill(warehouse, order.items):
        eligible.append(warehouse)

    if eligible is empty:
      raise DomainError("No warehouse can fulfill order " + order.id)

    // Prefer the warehouse closest to the customer (domain rule)
    chosen = self.selectNearest(eligible, order.shippingAddress)

    return AllocationPlan(
      orderId:     order.id,
      warehouseId: chosen.id,
      items:       order.items
    )

  // Check if a single warehouse has sufficient stock for all items
  private canFulfill(warehouse: Warehouse, items: List<OrderItem>) → Boolean:
    for item in items:
      stock = warehouse.stockFor(item.productId)
      if stock < item.quantity:
        return false
    return true

  // Select the warehouse with the shortest distance to the address
  private selectNearest(warehouses: List<Warehouse>, address: Address) → Warehouse:
    return warehouses.minBy(w → w.distanceTo(address))
}
```

**What it does NOT do:**
- Does not query a database for warehouses or stock
- Does not update stock levels (that happens after allocation, via a use case)
- Does not call any external API
- Does not know about delivery or transport protocols

---

## How domain services are used

Domain services are called by the **application layer** (use cases), not by
other domain objects or adapters.

```
// In AllocateInventoryUseCase (application layer — NOT here):
UseCase AllocateInventoryUseCase {
  execute(command: AllocateInventoryCommand):
    order      = orderRepository.findById(command.orderId)
    warehouses = warehouseRepository.findAll()

    // Call the domain service with pre-loaded domain objects
    plan = inventoryAllocator.allocate(order, warehouses)

    allocationRepository.save(plan)
}
```

The domain service receives fully-constructed domain objects. It never loads
data itself — the use case handles that through outbound ports.

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Stateless | No instance variables; all state comes from parameters |
| No I/O | Neither service touches a DB, API, or file system |
| Pure domain types | All parameters and return types are domain types |
| Multi-entity logic | Spans `Cart`+`Discount`, or `Order`+`Warehouse`+`Stock` |
| Application calls service | Use case loads data via repositories, then passes to service |
| Domain errors | Invalid inputs raise `DomainError`, not HTTP errors |
