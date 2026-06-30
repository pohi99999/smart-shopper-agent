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
import { useSubscription, PRODUCT_IDS } from '../context/SubscriptionContext';

// ── Types ─────────────────────────────────────────────────────────────────────

interface Props {
  onClose: () => void;
}

// ── Feature list data ─────────────────────────────────────────────────────────

interface Feature {
  icon: string;
  title: string;
  description: string;
}

const FEATURES: Feature[] = [
  {
    icon: '♾️',
    title: 'Korlátlan optimalizálás',
    description:
      'Naponta tetszőleges számú bevásárlólista feldolgozása az AI-motorral.',
  },
  {
    icon: '🚫',
    title: 'Teljesen reklámmentes',
    description: 'Élvezd az appot zavaró hirdetések és megszakítások nélkül.',
  },
  {
    icon: '📊',
    title: 'Intelligens árhistória',
    description:
      'Interaktív diagramok a termékek hetenkénti áringadozásairól.',
  },
  {
    icon: '🔔',
    title: 'Ár-riasztások',
    description:
      'Értesítés, ha a kedvenc terméked elér egy általad megadott célárát.',
  },
  {
    icon: '📋',
    title: 'Lista kedvencek',
    description: 'Bevásárlólisták mentése és egyetlen koppintással visszatöltése.',
  },
  {
    icon: '🏪',
    title: 'Több bolt, jobb ár',
    description:
      'Kibővített boltlista: Lidl, Spar, Tesco és még több lánc összehasonlítása.',
  },
];

// ── Component ─────────────────────────────────────────────────────────────────

export default function PaywallScreen({ onClose }: Props) {
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
        accessibilityLabel="Bezárás"
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
            Vásárolj okosabban. Spórolj többet.
          </Text>
        </View>

        {/* Feature list */}
        <View style={styles.featuresCard}>
          {FEATURES.map((f, i) => (
            <View
              key={f.title}
              style={[styles.featureRow, i < FEATURES.length - 1 && styles.featureRowBorder]}
            >
              <Text style={styles.featureIcon}>{f.icon}</Text>
              <View style={styles.featureText}>
                <Text style={styles.featureTitle}>{f.title}</Text>
                <Text style={styles.featureDescription}>{f.description}</Text>
              </View>
            </View>
          ))}
        </View>

        {/* Pricing */}
        <View style={styles.pricingRow}>
          <PricingBadge
            label="Havi"
            price="990 Ft"
            period="/hó"
            highlighted={false}
          />
          <PricingBadge
            label="Éves"
            price="7 990 Ft"
            period="/év"
            highlighted
            savingsTag="32% megtakarítás"
          />
        </View>

        {/* CTA */}
        {isPro ? (
          <View style={styles.successBanner}>
            <Text style={styles.successText}>✅ Pro fiók aktiválva!</Text>
          </View>
        ) : (
          <TouchableOpacity
            style={[styles.ctaButton, isLoading && styles.ctaButtonDisabled]}
            onPress={handleSubscribe}
            disabled={isLoading}
            activeOpacity={0.85}
            accessibilityLabel="Pro előfizetés indítása"
            accessibilityRole="button"
          >
            {isLoading ? (
              <ActivityIndicator color="#FFFFFF" />
            ) : (
              <>
                <Text style={styles.ctaText}>Előfizetés indítása</Text>
                <Text style={styles.ctaSubText}>Bármikor lemondható</Text>
              </>
            )}
          </TouchableOpacity>
        )}

        {/* Restore */}
        <TouchableOpacity
          onPress={handleRestore}
          disabled={isLoading}
          style={styles.restoreButton}
          accessibilityLabel="Korábbi vásárlás visszaállítása"
        >
          <Text style={styles.restoreText}>Korábbi vásárlás visszaállítása</Text>
        </TouchableOpacity>

        {/* Legal */}
        <Text style={styles.legalText}>
          Az előfizetés automatikusan megújul, hacsak a megújítás előtt legalább
          24 órával le nem mondod. Az iTunes-fiókodból kerül a kifizetés a
          megerősítéskor.
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
