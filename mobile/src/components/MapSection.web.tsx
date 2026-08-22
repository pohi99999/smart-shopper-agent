import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { OptimizeResponse, Coordinate } from '../services/api';

interface MapSectionProps {
  coords: Coordinate | null;
  result: OptimizeResponse | null;
}

// react-native-maps has no web renderer (it falls back to a native view that
// does not exist in the browser), so the web build uses this stop list
// instead of crashing. The interactive map stays native-only (see
// MapSection.tsx); ResultSection already lists the same stops with items.
export function MapSection({ coords, result }: MapSectionProps) {
  return (
    <View style={styles.mapCard}>
      <Text style={styles.cardTitle}>Térkép</Text>
      <Text style={styles.webNotice}>
        Az interaktív térkép jelenleg csak a mobilalkalmazásban érhető el.
      </Text>
      {coords && (
        <Text style={styles.coordLine}>
          Saját helyzet: {coords.latitude.toFixed(4)}, {coords.longitude.toFixed(4)}
        </Text>
      )}
      {result &&
        result.route_plan.steps.map((step, index) => (
          <Text key={index} style={styles.coordLine}>
            {index + 1}. {step.shop_name}
            {step.coordinates
              ? ` (${step.coordinates.latitude.toFixed(4)}, ${step.coordinates.longitude.toFixed(4)})`
              : ''}
          </Text>
        ))}
    </View>
  );
}

const styles = StyleSheet.create({
  mapCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 18,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.08,
    shadowRadius: 12,
    elevation: 3,
    marginBottom: 24,
  },
  cardTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1C1C1E',
    marginBottom: 12,
  },
  webNotice: {
    fontSize: 14,
    color: '#8E8E93',
    marginBottom: 12,
  },
  coordLine: {
    fontSize: 14,
    color: '#3A3A3C',
    marginBottom: 6,
  },
});
