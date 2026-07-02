import React, { useCallback } from 'react';
import {
  ActivityIndicator,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { useSubscription, PRODUCT_IDS } from '../context/SubscriptionContext';

// ── Types ─────────────────────────────────────────────────────────────────────

interface Props {
  onClose: () => void;
}

// ── Feature list item interface ───────────────────────────────────────────────

interface FeatureItem {
  icon: string;
  key: string;
}

const FEATURE_ITEMS: FeatureItem[] = [
  { icon: '♾️', key: 'unlimited' },
  { icon: '🚫', key: 'noAds' },
  { icon: '📊', key: 'priceHistory' },
  { icon: '🔔', key: 'priceAlerts' },
  { icon: '📋', key: 'favorites' },
  { icon: '🏪', key: 'moreStores' },
];

// ── Component ─────────────────────────────────────────────────────────────────

export default function PaywallScreen({ onClose }: Props) {
  const { t } = useTranslation();
  const { subscribe, restore, isLoading, isPro } = useSubscription();

  const handleSubscribe = useCallback(async () => {
    await subscribe(PRODUCT_IDS.MONTHLY);
  }, [subscribe]);

  const handleRestore = useCallback(async () => {
    await restore();
  }, [restore]);

  // Auto-close after successful subscription
  React.useEffect(() => {
    if (isPro) {
      const timer = setTimeout(() => onClose(), 800);
      return () => clearTimeout(timer);
    }
  }, [isPro, onClose]);

  return (
    <View style={styles.root}>
      {/* Close button */}
      <TouchableOpacity
        style={styles.closeButton}
        onPress={onClose}
        accessibilityLabel={t('paywall.closeAccessibility')}
        accessibilityRole="button"
      >
        <Text style={styles.closeIcon}>✕</Text>
      </TouchableOpacity>

      <ScrollView
        contentContainerStyle={styles.scroll}
        showsVerticalScrollIndicator={false}
        bounces={Platform.OS !== 'web'}
      >
        {/* Hero */}
        <View style={styles.hero}>
          <View style={styles.crownBadge}>
            <Text style={styles.crownEmoji}>👑</Text>
          </View>
          <Text style={styles.heroTitle}>Smart Shopper</Text>
          <Text style={styles.heroPro}>PRO</Text>
          <Text style={styles.heroTagline}>
            {t('paywall.heroTagline')}
          </Text>
        </View>

        {/* Feature list */}
        <View style={styles.featuresCard}>
          {FEATURE_ITEMS.map((item, i) => (
            <View
              key={item.key}
              style={[styles.featureRow, i < FEATURE_ITEMS.length - 1 && styles.featureRowBorder]}
            >
              <Text style={styles.featureIcon}>{item.icon}</Text>
              <View style={styles.featureText}>
                <Text style={styles.featureTitle}>
                  {t(`paywall.features.${item.key}.title`)}
                </Text>
                <Text style={styles.featureDescription}>
                  {t(`paywall.features.${item.key}.description`)}
                </Text>
              </View>
            </View>
          ))}
        </View>

        {/* Pricing */}
        <View style={styles.pricingRow}>
          <PricingBadge
            label={t('paywall.pricing.monthly')}
            price="990 Ft"
            period={t('paywall.pricing.monthlyPeriod')}
            highlighted={false}
          />
          <PricingBadge
            label={t('paywall.pricing.yearly')}
            price="7 990 Ft"
            period={t('paywall.pricing.yearlyPeriod')}
            highlighted
            savingsTag={t('paywall.pricing.savingsTag')}
          />
        </View>

        {/* CTA */}
        {isPro ? (
          <View style={styles.successBanner}>
            <Text style={styles.successText}>{t('paywall.cta.proActive')}</Text>
          </View>
        ) : (
          <TouchableOpacity
            style={[styles.ctaButton, isLoading && styles.ctaButtonDisabled]}
            onPress={handleSubscribe}
            disabled={isLoading}
            activeOpacity={0.85}
            accessibilityLabel={t('paywall.cta.accessibilityLabel')}
            accessibilityRole="button"
          >
            {isLoading ? (
              <ActivityIndicator color="#FFFFFF" />
            ) : (
              <>
                <Text style={styles.ctaText}>{t('paywall.cta.subscribe')}</Text>
                <Text style={styles.ctaSubText}>{t('paywall.cta.cancelAnytime')}</Text>
              </>
            )}
          </TouchableOpacity>
        )}

        {/* Restore */}
        <TouchableOpacity
          onPress={handleRestore}
          disabled={isLoading}
          style={styles.restoreButton}
          accessibilityLabel={t('paywall.restore.button')}
        >
          <Text style={styles.restoreText}>{t('paywall.restore.button')}</Text>
        </TouchableOpacity>

        {/* Legal */}
        <Text style={styles.legalText}>
          {t('paywall.legal')}
        </Text>
      </ScrollView>
    </View>
  );
}

// ── Sub-component ─────────────────────────────────────────────────────────────

interface PricingBadgeProps {
  label: string;
  price: string;
  period: string;
  highlighted: boolean;
  savingsTag?: string;
}

function PricingBadge({ label, price, period, highlighted, savingsTag }: PricingBadgeProps) {
  return (
    <View style={[styles.pricingBadge, highlighted && styles.pricingBadgeHighlighted]}>
      {savingsTag && (
        <View style={styles.savingsTag}>
          <Text style={styles.savingsTagText}>{savingsTag}</Text>
        </View>
      )}
      <Text style={[styles.pricingLabel, highlighted && styles.pricingLabelHighlighted]}>
        {label}
      </Text>
      <Text style={[styles.pricingPrice, highlighted && styles.pricingPriceHighlighted]}>
        {price}
      </Text>
      <Text style={[styles.pricingPeriod, highlighted && styles.pricingPeriodHighlighted]}>
        {period}
      </Text>
    </View>
  );
}

// ── Styles ────────────────────────────────────────────────────────────────────

const GOLD = '#F5A623';
const GOLD_DARK = '#D4831A';
const BLUE = '#007AFF';
const BG = '#F2F2F7';
const CARD_BG = '#FFFFFF';
const TEXT_PRIMARY = '#1C1C1E';
const TEXT_SECONDARY = '#8E8E93';

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: BG,
  },
  closeButton: {
    position: 'absolute',
    top: Platform.OS === 'ios' ? 54 : 20,
    right: 20,
    zIndex: 10,
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: 'rgba(0,0,0,0.08)',
    alignItems: 'center',
    justifyContent: 'center',
  },
  closeIcon: {
    fontSize: 14,
    color: TEXT_PRIMARY,
    fontWeight: '600',
  },
  scroll: {
    paddingHorizontal: 20,
    paddingTop: Platform.OS === 'ios' ? 60 : 32,
    paddingBottom: 48,
  },

  // ── Hero ──
  hero: {
    alignItems: 'center',
    marginBottom: 28,
  },
  crownBadge: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: GOLD,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
    shadowColor: GOLD,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.35,
    shadowRadius: 16,
    elevation: 8,
  },
  crownEmoji: {
    fontSize: 34,
  },
  heroTitle: {
    fontSize: 28,
    fontWeight: '800',
    color: TEXT_PRIMARY,
    letterSpacing: 0.3,
  },
  heroPro: {
    fontSize: 28,
    fontWeight: '900',
    color: GOLD_DARK,
    letterSpacing: 6,
    marginTop: -4,
    marginBottom: 10,
  },
  heroTagline: {
    fontSize: 16,
    color: TEXT_SECONDARY,
    textAlign: 'center',
  },

  // ── Features ──
  featuresCard: {
    backgroundColor: CARD_BG,
    borderRadius: 18,
    paddingVertical: 4,
    marginBottom: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.07,
    shadowRadius: 12,
    elevation: 3,
  },
  featureRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingVertical: 14,
    paddingHorizontal: 18,
  },
  featureRowBorder: {
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#E5E5EA',
  },
  featureIcon: {
    fontSize: 22,
    marginRight: 14,
    marginTop: 1,
  },
  featureText: {
    flex: 1,
  },
  featureTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: TEXT_PRIMARY,
    marginBottom: 2,
  },
  featureDescription: {
    fontSize: 13,
    color: TEXT_SECONDARY,
    lineHeight: 18,
  },

  // ── Pricing ──
  pricingRow: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  pricingBadge: {
    flex: 1,
    backgroundColor: CARD_BG,
    borderRadius: 16,
    padding: 16,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 8,
    elevation: 2,
    borderWidth: 1.5,
    borderColor: 'transparent',
    position: 'relative',
    overflow: 'visible',
  },
  pricingBadgeHighlighted: {
    borderColor: GOLD,
    backgroundColor: '#FFFBF0',
  },
  savingsTag: {
    position: 'absolute',
    top: -12,
    backgroundColor: GOLD,
    borderRadius: 8,
    paddingHorizontal: 8,
    paddingVertical: 2,
  },
  savingsTagText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  pricingLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: TEXT_SECONDARY,
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: 4,
    marginTop: 8,
  },
  pricingLabelHighlighted: {
    color: GOLD_DARK,
  },
  pricingPrice: {
    fontSize: 22,
    fontWeight: '800',
    color: TEXT_PRIMARY,
  },
  pricingPriceHighlighted: {
    color: GOLD_DARK,
  },
  pricingPeriod: {
    fontSize: 12,
    color: TEXT_SECONDARY,
  },
  pricingPeriodHighlighted: {
    color: GOLD_DARK,
  },

  // ── CTA ──
  ctaButton: {
    backgroundColor: GOLD,
    borderRadius: 16,
    paddingVertical: 18,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
    shadowColor: GOLD,
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.3,
    shadowRadius: 12,
    elevation: 5,
  },
  ctaButtonDisabled: {
    opacity: 0.7,
  },
  ctaText: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: '700',
    letterSpacing: 0.3,
  },
  ctaSubText: {
    color: 'rgba(255,255,255,0.85)',
    fontSize: 12,
    marginTop: 3,
  },
  successBanner: {
    backgroundColor: '#34C759',
    borderRadius: 16,
    paddingVertical: 18,
    alignItems: 'center',
    marginBottom: 16,
  },
  successText: {
    color: '#FFFFFF',
    fontSize: 17,
    fontWeight: '700',
  },

  // ── Restore / Legal ──
  restoreButton: {
    alignItems: 'center',
    marginBottom: 20,
  },
  restoreText: {
    color: BLUE,
    fontSize: 14,
    fontWeight: '500',
  },
  legalText: {
    fontSize: 11,
    color: TEXT_SECONDARY,
    textAlign: 'center',
    lineHeight: 16,
  },
});
