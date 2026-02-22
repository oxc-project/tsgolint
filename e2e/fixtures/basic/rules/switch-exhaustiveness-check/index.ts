// Examples of incorrect code for switch-exhaustiveness-check rule

type OrderStatus = 'pending' | 'approved' | 'rejected';

function handleStatus(status: OrderStatus) {
  switch (status) {
    case 'pending':
      return 'Waiting for approval';
    case 'approved':
      return 'Request approved';
    // Missing 'rejected' case
  }
}