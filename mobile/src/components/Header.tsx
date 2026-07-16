import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

interface HeaderProps {
  isPro: boolean;
  onShowPaywall?: () => void;
}

export function Header({ isPro, onShowPaywall }: HeaderProps) {
  return (
    <View style={styles.header}>
      <View style={styles.headerRow}>
        <View>
          <Text style={styles.title}>Smart Shopper</Text>
          <Text style={styles.subtitle}>Személyes bevásárló asszisztens</Text>
        </View>
        {!isPro && onShowPaywall && (
          <TouchableOpacity
            style={styles.proButton}
            onPress={onShowPaywall}
            activeOpacity={0.8}
            accessibilityLabel="Smart Shopper Pro megnyitása"
            accessibilityRole="button"
          >
            <Text style={styles.proButtonIcon}>👑</Text>
            <Text style={styles.proButtonText}>Go Pro</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  header: {
    marginBottom: 24,
    marginTop: 10,
  },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  title: {
    fontSize: 34,
    fontWeight: '800',
    color: '#1C1C1E',
    letterSpacing: 0.37,
  },
  subtitle: {
    fontSize: 16,
    color: '#8E8E93',
    marginTop: 4,
  },
  proButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#F5A623',
    borderRadius: 20,
    paddingHorizontal: 14,
    paddingVertical: 8,
    shadowColor: '#F5A623',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.25,
    shadowRadius: 6,
    elevation: 3,
  },
  proButtonIcon: {
    fontSize: 14,
    marginRight: 5,
  },
  proButtonText: {
    color: '#FFFFFF',
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0.3,
  },
});
