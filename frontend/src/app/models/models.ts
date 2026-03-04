// ── Auth ──

export interface AuthResponse {
    token: string;
    user: UserResponse;
}

export interface UserResponse {
    id: string;
    email: string;
}

// ── Accounts ──

export interface Account {
    item_id: string;
    account_id: string;
    account_name: string;
    account_type: string;
    account_subtype?: string;
    current_balance: string;
    available_balance: string;
    iso_currency_code: string;
}

// ── Transactions ──

export interface Transaction {
    transaction_id: string;
    account_id: string;
    date: string;
    name: string;
    amount: string;
    pending: boolean;
    merchant_name?: string;
    logo_url?: string;
    personal_finance_category?: string;
    detailed_category?: string;
    category_confidence_level?: string;
    category_icon_url?: string;
    created_at: string;
    account_name?: string;
}

// ── Budgets ──

export interface Budget {
    id: string;
    user_id: string;
    category: string;
    limit_amount: string;
    amount_spent: string;
    period: string;
    start_date: string;
    end_date: string | null;
    created_at: string;
}

export interface CreateBudgetRequest {
    category: string;
    limit_amount: string;
    period: string;
    start_date: string;
    end_date?: string;
}

// ── Plaid ──

export interface CreateLinkTokenResponse {
    link_token: string;
}

export interface ExchangeTokenRequest {
    public_token: string;
    institution_name: string;
    account_name: string;
    account_type: string;
}

export interface ExchangeTokenResponse {
    account_id: string;
    item_id: string;
}
