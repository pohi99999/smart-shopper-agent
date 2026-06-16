import React, { useState } from 'react';
import {
  StyleSheet,
  Text,
  View,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  ScrollView,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import { optimizeShoppingRoute, OptimizeResponse } from '../services/api';

export default function ShoppingListScreen() {
  const [inputText, setInputText] = useState('Kérek 10 tojást és 1 kenyeret.');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<OptimizeResponse | null>(null);

  const handleOptimize = async () => {
    if (!inputText.trim()) {
      Alert.alert('Hiba', 'Kérlek írd be a bevásárlólistádat!');
      return;
    }

    setLoading(true);
    setResult(null);

    try {
      // Budapest központ koordináták
      const lat = 47.4979;
      const lon = 19.0402;
      const response = await optimizeShoppingRoute(inputText, lat, lon);
      setResult(response);
    } catch (error) {
      Alert.alert(
        'Sikertelen optimalizálás',
        'Nem sikerült csatlakozni az optimalizáló szerverhez. Ellenőrizd, hogy fut-e a backend!'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.container}
      >
        <ScrollView contentContainerStyle={styles.scrollContainer} keyboardShouldPersistTaps="handled">
          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.title}>Smart Shopper</Text>
            <Text style={styles.subtitle}>Személyes bevásárló asszisztens</Text>
          </View>

          {/* Input Section */}
          <View style={styles.card}>
            <Text style={styles.cardTitle}>Mit szeretnél vásárolni?</Text>
            <TextInput
              style={styles.textInput}
              multiline
              numberOfLines={4}
              value={inputText}
              onChangeText={setInputText}
              placeholder="Írd be a listát szabad szöveggel..."
              placeholderTextColor="#999"
            />
            
            {loading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="large" color="#007AFF" />
                <Text style={styles.loadingText}>Útvonal és árak optimalizálása...</Text>
              </View>
            ) : (
              <TouchableOpacity style={styles.button} onPress={handleOptimize} activeOpacity={0.8}>
                <Text style={styles.buttonText}>Útvonal Optimalizálása</Text>
              </TouchableOpacity>
            )}
          </View>

          {/* Result Section */}
          {result && (
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
          )}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#F2F2F7', // iOS Light background
  },
  container: {
    flex: 1,
  },
  scrollContainer: {
    padding: 20,
    paddingBottom: 40,
  },
  header: {
    marginBottom: 24,
    marginTop: 10,
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
  card: {
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
  textInput: {
    backgroundColor: '#F2F2F7',
    borderRadius: 12,
    padding: 14,
    fontSize: 16,
    color: '#1C1C1E',
    minHeight: 100,
    textAlignVertical: 'top',
    marginBottom: 16,
  },
  button: {
    backgroundColor: '#007AFF', // Apple Blue
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 2,
  },
  buttonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 10,
  },
  loadingText: {
    marginTop: 10,
    color: '#8E8E93',
    fontSize: 14,
  },
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
