import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { OptimizeResponse } from '../services/api';

interface ResultSectionProps {
  result: OptimizeResponse | null;
}

export function ResultSection({ result }: ResultSectionProps) {
  if (!result) return null;

  return (
    <View style={styles.resultContainer}>
      <View style={styles.summaryCard}>
        <Text style={styles.summaryLabel}>Becsült végösszeg</Text>
        <Text style={styles.summaryCost}>
          {result.total_cost.toLocaleString('hu-HU')} Ft
        </Text>
      </View>

      <Text style={styles.sectionTitle}>Optimális útiterv</Text>

      {result.route_plan.steps.map((step, index) => (
        <View key={index} style={styles.stepCard}>
          <View style={styles.stepHeader}>
            <View style={styles.stepBadge}>
              <Text style={styles.stepBadgeText}>{index + 1}</Text>
            </View>
            <Text style={styles.shopName}>{step.shop_name}</Text>
          </View>
          <View style={styles.stepBody}>
            <Text style={styles.itemsLabel}>Megvásárolandó tételek:</Text>
            {step.items.map((item, itemIdx) => (
              <View key={itemIdx} style={styles.itemRow}>
                <Text style={styles.itemBullet}>•</Text>
                <Text style={styles.itemText}>{item}</Text>
              </View>
            ))}
          </View>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  resultContainer: {
    marginTop: 8,
  },
  summaryCard: {
    backgroundColor: '#34C759', // Apple Green
    borderRadius: 16,
    padding: 20,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#34C759',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.15,
    shadowRadius: 10,
    elevation: 4,
    marginBottom: 24,
  },
  summaryLabel: {
    color: 'rgba(255, 255, 255, 0.85)',
    fontSize: 14,
    fontWeight: '500',
    textTransform: 'uppercase',
    letterSpacing: 1.2,
    marginBottom: 4,
  },
  summaryCost: {
    color: '#FFFFFF',
    fontSize: 32,
    fontWeight: '800',
  },
  sectionTitle: {
    fontSize: 22,
    fontWeight: '700',
    color: '#1C1C1E',
    marginBottom: 14,
  },
  stepCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 8,
    elevation: 2,
    borderLeftWidth: 4,
    borderLeftColor: '#007AFF',
  },
  stepHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  stepBadge: {
    backgroundColor: '#007AFF',
    width: 24,
    height: 24,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 10,
  },
  stepBadgeText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '700',
  },
  shopName: {
    fontSize: 18,
    fontWeight: '700',
    color: '#1C1C1E',
  },
  stepBody: {
    paddingLeft: 34,
  },
  itemsLabel: {
    fontSize: 14,
    color: '#8E8E93',
    marginBottom: 6,
    fontWeight: '500',
  },
  itemRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  itemBullet: {
    fontSize: 16,
    color: '#007AFF',
    marginRight: 6,
  },
  itemText: {
    fontSize: 15,
    color: '#3A3A3C',
  },
});
