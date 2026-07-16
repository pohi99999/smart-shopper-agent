import React from 'react';
import {
  StyleSheet,
  ScrollView,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useShoppingOptimizer } from '../hooks/useShoppingOptimizer';
import { useSubscription } from '../context/SubscriptionContext';
import { Header } from '../components/Header';
import { InputSection } from '../components/InputSection';
import { MapSection } from '../components/MapSection';
import { ResultSection } from '../components/ResultSection';

interface Props {
  onShowPaywall?: () => void;
}

export default function ShoppingListScreen({ onShowPaywall }: Props) {
  const {
    inputText,
    setInputText,
    loading,
    result,
    coords,
    handleOptimize,
  } = useShoppingOptimizer();

  const { isPro } = useSubscription();

  return (
    <SafeAreaView style={styles.safeArea}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.container}
      >
        <ScrollView contentContainerStyle={styles.scrollContainer} keyboardShouldPersistTaps="handled">
          <Header isPro={isPro} onShowPaywall={onShowPaywall} />
          <InputSection
            inputText={inputText}
            setInputText={setInputText}
            loading={loading}
            handleOptimize={handleOptimize}
          />
          <MapSection coords={coords} result={result} />
          <ResultSection result={result} />
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
});
