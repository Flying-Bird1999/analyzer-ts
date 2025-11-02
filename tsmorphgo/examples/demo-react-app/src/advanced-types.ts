// Advanced TypeScript Types for QuickInfo Testing

// 1. Conditional Types
type IsString<T> = T extends string ? "yes" : "no";
type A = IsString<string>; // "yes"
type B = IsString<number>; // "no"

// 2. Mapped Types
type Readonly<T> = {
    readonly [P in keyof T]: T[P];
};

interface WritableUser {
    id: number;
    name: string;
}

type ReadonlyUser = Readonly<WritableUser>;

// 3. Template Literal Types
type World = "world";
type Greeting = `hello ${World}`;

// 4. Nested and Complex Interfaces
export interface ApiResponse<T> {
    data: T;
    status: 'success' | 'error';
    meta: {
        requestId: string;
        timestamp: Date;
        pagination?: {
            currentPage: number;
            totalPages: number;
            totalItems: number;
        };
    };
}

export interface Product {
    id: string;
    name: string;
    price: number;
    details: ProductDetails;
}

export interface ProductDetails {
    description: string;
    specifications: Record<string, string>;
    reviews: Review[];
}

export interface Review {
    author: string;
    rating: number;
    comment: string;
}

// 5. Complex Function with Generics and Callbacks
export function processData<T, U>(
    data: T[],
    processor: (item: T) => U,
    onComplete: (result: U[]) => void
): void {
    const result = data.map(processor);
    onComplete(result);
}
