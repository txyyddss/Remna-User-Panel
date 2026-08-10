# Platform architecture

Platform packages expose small infrastructure contracts to application and domain services. They own resource lifecycles, cancellation, storage primitives, and cryptographic mechanics; domain validation and HTTP policy remain outside this layer. Long-running workers stop through context cancellation before dependent resources close.
