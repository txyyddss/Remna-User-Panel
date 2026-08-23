# Schema shards

Schema definitions are grouped into bounded sequential shards. The root component registry references each definition by JSON Pointer.

- `schemas-01.yaml`: `ID`, `Timestamp`, `DecimalInteger`, `ApiError`, `TelegramAuthRequest`, `OnboardingStep`, `User`, `AuthState`, `JoinInvite`, `OnboardingInvites`, `MembershipState`, `UsernameRequest`, `AgreementRequest`, `WelcomeMessage`, `OnboardingAgreement`, `LocalizedOnboardingContent`, `PublishedOnboarding`, `OnboardingBundle`
- `schemas-02.yaml`: `OnboardingDraftWrite`, `OnboardingPublishRequest`, `Money`, `ResetStrategy`, typed squad profiles, `SquadProduct` with live accessible nodes, `Combo`, nullable-provider `CatalogNode`, `Catalog`, `ComboWrite`, `SquadProductWrite`
- `schemas-03.yaml`: `RemnaNode`, `PurchaseRequest`, `PurchaseQuote`, `EntitlementStatus`, `Entitlement`, `Purchase`, `SubscriptionRotation`, `TopNode`, `Statistics`, `StatisticPoint`
- `schemas-04.yaml`: `StatisticSlice`, `AdminStatistics`, `Dashboard`, `LedgerKind`, `LedgerEntry`, `PageInfo`, `LedgerPage`, `PaymentMethodID`, `PaymentOrderMethodID`, `PaymentProvider`, `PaymentStatus`, `PaymentOrderRequest`
- `schemas-05.yaml`: `PaymentOrder`, `ActivityGame`, `ActivityGameWrite`, `ActivityGameList`, `LuckyDraw`, `NoPrizeReward`, `TXBDeltaReward`, `CouponGrantReward`, `SubscriptionExtensionReward`, `Reward`, `ActivityResult`
- `schemas-06.yaml`: `GroupMessageRewardStatus`, `ActivityOverview`, `ActivityBetRequest`, `ActivityDrawRequest`, `ActivitySettings`, `ActivitySettingsWrite`, `LuckyDrawPrize`, `LuckyDrawPrizeWrite`, `LuckyDrawAdmin`, `LuckyDrawWrite`, `LuckyDrawAdminList`, `CouponKind`, `CouponDiscountMode`
- `schemas-07.yaml`: `Coupon`, `CouponWrite`, `CouponList`, `CouponGrant`, `CouponGrantList`, `CouponRedeemRequest`, `CouponRedemption`, `QuestionnaireStatus`, `Questionnaire`, `QuestionnaireWrite`, `QuestionnaireList`
- `schemas-08.yaml`: `QuestionnaireParticipation`, `ActiveQuestionnaire`, `QuestionnaireHistoryItem`, `QuestionnaireHistoryList`, `QuestionnaireImportStatus`, `QuestionnaireImportAnalysis`, `QuestionnaireImportPreview`, `QuestionnaireImportAnalyzeRequest`, `QuestionnaireSettlementReport`, `QuestionnaireImportState`, `EmbyRating`, `EmbyLibrary`, `EmbyAccount`
- `schemas-09.yaml`: `EmbyOverview`, `EmbyPreferencesRequest`, `EmbySetupRequest`, `EmbyPasswordRequest`, `EmbyAccountList`, `AdminSetting`, `AdminSettingCreate`, `AdminSettingUpdate`, `AdminUserSummary`, `AdminUserPage`, `BalanceAdjustmentRequest`, `ReasonRequest`, `EntitlementPage`, `AdminEntitlement`
- `schemas-10.yaml`: `PaymentReturnStatus`, `PaymentPage`, `AdminPaymentOrder`, `RefundRequest`, `Refund`, `RefundPage`, `DatabaseColumn`, `DatabaseTable`, `DatabaseTableList`, `DatabaseBlobValue`, `DatabaseValue`, `DatabaseKey`, `DatabaseValues`, `DatabaseRow`, `DatabaseRowPage`, `DatabaseFilterOperator`
- `schemas-11.yaml`: `DatabaseFilter`, `DatabaseQueryRequest`, `DatabaseMutationAction`, `DatabaseMutationReviewRequest`, `DatabaseMutationApplyRequest`, `DatabaseCompatibilityUpdate`, `DatabaseMutationReview`, `DatabaseMutationResult`, `RestoreRequest`, `RestoreOperation`, `BackupRun`, `JobStatus`, `Job`, `JobPage`
- `schemas-12.yaml`: `AuditEvent`, `AuditPage`, `TelegramUpdate`, `FlexibleDecimal`, `BEPusdtNotification`, `BEPusdtUnsignedNotification`, `HealthStatus`, `RuntimeReadiness`
- `schemas-13.yaml`: `CourtesyCredit`
- `schemas-14.yaml`: `DashboardNodeUsage`
- `schemas-15.yaml`: `AutoRenewal`, `AutoRenewalUpdate`, `AutoRenewalFailure`, stable-ID `PaymentProfile`, and `PaymentProfileWrite`
- `schemas-16.yaml`: live `RolloverWindow` and `RolloverProjection` aggregate response schemas
- `schemas-17.yaml`: durable operation and payment-operation receipts, signed connection projections, active IP blocks, account-wide reset automation, reset/refund quotes, and billing amount limits
- `schemas-18.yaml`: admin entitlement/combo mutations, minute-normalized bulk extension requests with a legacy-days compatibility input, and aggregate user detail including active IP blocks
- `schemas-19.yaml`: cached remote/database statistics including live predicted rollover, normalized distributions, and node Geocheck image contracts
- `schemas-20.yaml`: affiliate member projections, tier rewards, bot identity, and optimistic administrator writes
- `schemas-21.yaml`: `PaymentMethod` with provider display name and pending-order-aware `BalanceResponse`
- `schemas-22.yaml`: revisioned node-compensation configuration, outage projections, review writes, and recipient-safe event pages
