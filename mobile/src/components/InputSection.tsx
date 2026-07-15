import React from 'react';
import { View, Text, TextInput, TouchableOpacity, ActivityIndicator, StyleSheet } from 'react-native';

interface InputSectionProps {
  inputText: string;
  setInputText: (text: string) => void;
  loading: boolean;
  handleOptimize: () => void;
}

export function InputSection({
  inputText,
  setInputText,
  loading,
  handleOptimize,
}: InputSectionProps) {
  return (
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
  );
}

const styles = StyleSheet.create({
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
});
