// 1. Pessimistic locking (SELECT FOR UPDATE), uses long-lived row locks, held until commit/rollback.
//
// BEGIN;
// SELECT * FROM orders WHERE id = 1 FOR UPDATE;
// -- row is locked
// UPDATE orders SET status = 'paid' WHERE id = 1;
// COMMIT;
//
// 2. Optimistic Locking vs Pessimistic Locking
// | Aspect            | Optimistic Locking | Pessimistic Locking |
// | ----------------- | ------------------ | ------------------- |
// | Locks             | ❌ No locks         | ✅ Row locks       |
// | Conflict handling | Detect after       | Prevent before      |
// | Performance       | High               | Lower               |
// | Deadlocks         | Impossible         | Possible            |
// | Use case          | High throughput and| Frequent conflicts  |
// |                   | Rare conflicts     | Strict ordering     | (Financial trasfers requires)
// | *Blocking         | non-blocking       | blocking            |
//
//   - Pessimistic locking | Optimistic locking
//     T1 locks row A       | T1 update succeeds
//     T2 waits             | T2 update fails immediately
//     T3 waits             | T3 update fails immediately
//     ...                  | ...
//
// 3. Pessimistic may cause deadlock, a classic example is two transactions try to transfer money in opposite directions.
// Transaction 1:                                      | Transaction 2:
// BEGIN;                                              | BEGIN;
// SELECT * FROM accounts WHERE id = 1 FOR UPDATE;     | SELECT * FROM accounts WHERE id = 2 FOR UPDATE；
// -- locks row 1                                      | -- locks row2
// SELECT * FROM accounts WHERE id = 2 FOR UPDATE;     | SELECT * FROM accounts WHERE id = 1 FROM UPDATE;
// -- waits for row 2                                  | -- wait for row 1
//
// So if must use pessimistic locking, always lock in the same order (e.g. always lock smaller ID first).
// Databases detect deadlocks automatically. One transaction is chosen as the victim.
// Deadlocks are expensive, when it happens:
// 1. Transactions are blocked waiting, threads pile up, connection pool exhausts and latency spikes => cascading failures possible
// 2. One transaction is forcefully rolled back, work already done is wasted
//
// 4. Atomic conditional update
// res := db.Exec(`
//     UPDATE seats
//     SET reserved = reserved + 1
//     WHERE id = ? AND reserved < capacity
// `, seatID)

//	if res.RowsAffected == 0 {
//	    return errors.New("no seats available")
//	}
//
// This statement is atomic because:
// - The condition check (reserved < capacity)
// - And the state change (reserved = reserved + 1)
// happen as one indivisible operation inside the database.
// There is no gap where another transaction can sneak in.
//
// This pattern is simpler and safer for counters and quotas.
package advanced
