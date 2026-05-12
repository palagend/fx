<template>
  <div class="calculator-container">
    <div class="card">
      <h2 class="page-title">
        <Icon icon="mdi:calculator" />
        <span>实用计算器</span>
      </h2>

      <div class="tabs">
        <button class="tab-btn" :class="{ active: activeTab === 'basic' }" @click="activeTab = 'basic'">
          <Icon icon="mdi:calculator" />
          <span>基础计算</span>
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'unit' }" @click="activeTab = 'unit'">
          <Icon icon="mdi:swap-horizontal" />
          <span>单位转换</span>
        </button>
      </div>

      <div v-if="activeTab === 'basic'" class="calculator">
        <div class="display">
          <div class="history">{{ history }}</div>
          <div class="input">{{ input || '0' }}</div>
        </div>

        <div class="buttons">
          <button class="btn btn-function" @click="backspace">
            <Icon icon="mdi:backspace" />
          </button>
          <button class="btn btn-clear" @click="clear">
            {{ clearButtonText }}
          </button>
          <button class="btn btn-operator" @click="append('%')">%</button>
          <button class="btn btn-operator" @click="append('÷')">
            <Icon icon="mdi:divide" />
          </button>

          <button class="btn btn-number" @click="append('7')">7</button>
          <button class="btn btn-number" @click="append('8')">8</button>
          <button class="btn btn-number" @click="append('9')">9</button>
          <button class="btn btn-operator" @click="append('×')">
            <Icon icon="mdi:multiply" />
          </button>

          <button class="btn btn-number" @click="append('4')">4</button>
          <button class="btn btn-number" @click="append('5')">5</button>
          <button class="btn btn-number" @click="append('6')">6</button>
          <button class="btn btn-operator" @click="append('-')">
            <Icon icon="mdi:minus" />
          </button>

          <button class="btn btn-number" @click="append('1')">1</button>
          <button class="btn btn-number" @click="append('2')">2</button>
          <button class="btn btn-number" @click="append('3')">3</button>
          <button class="btn btn-operator" @click="append('+')">
            <Icon icon="mdi:plus" />
          </button>

          <button class="btn btn-function" @click="toggleSign">+/-</button>
          <button class="btn btn-number" @click="append('0')">0</button>
          <button class="btn btn-number" @click="append('.')">
            <Icon icon="mdi:dot" />
          </button>
          <button class="btn btn-equals" @click="calculate">
            <Icon icon="mdi:equal" />
          </button>
        </div>
      </div>

      <div v-if="activeTab === 'unit'" class="unit-converter">
        <!-- iOS风格分段控制器 -->
        <div class="unit-type-segmented">
          <button 
            v-for="type in unitTypes" 
            :key="type.value"
            class="segment-btn"
            :class="{ active: unitType === type.value }"
            @click="setUnitType(type.value)"
          >
            <Icon :icon="type.icon" />
            <span>{{ type.label }}</span>
          </button>
        </div>

        <!-- 输入区域 -->
        <div class="unit-input-section">
          <div class="unit-display from" @click="showFromPicker = true">
            <div class="unit-value">{{ unitValue || 0 }}</div>
            <div class="unit-name">{{ getUnitLabel(fromUnit) }}</div>
          </div>
          
          <button class="unit-swap-btn" @click="swapUnits">
            <Icon icon="mdi:swap-vertical" />
          </button>
          
          <div class="unit-display to">
            <div class="unit-value">{{ convertedValue }}</div>
            <div class="unit-name">{{ getUnitLabel(toUnit) }}</div>
          </div>
        </div>

        <!-- 数字键盘 -->
        <div class="unit-keyboard">
          <button class="unit-key" @click="appendUnitValue('7')">7</button>
          <button class="unit-key" @click="appendUnitValue('8')">8</button>
          <button class="unit-key" @click="appendUnitValue('9')">9</button>
          <button class="unit-key unit-key-function" @click="backspaceUnitValue">
            <Icon icon="mdi:backspace" />
          </button>

          <button class="unit-key" @click="appendUnitValue('4')">4</button>
          <button class="unit-key" @click="appendUnitValue('5')">5</button>
          <button class="unit-key" @click="appendUnitValue('6')">6</button>
          <button class="unit-key unit-key-function" @click="clearUnitValue">C</button>

          <button class="unit-key" @click="appendUnitValue('1')">1</button>
          <button class="unit-key" @click="appendUnitValue('2')">2</button>
          <button class="unit-key" @click="appendUnitValue('3')">3</button>
          <button class="unit-key unit-key-function" @click="toggleUnitSign">+/-</button>

          <button class="unit-key unit-key-zero" @click="appendUnitValue('0')">0</button>
          <button class="unit-key" @click="appendUnitValue('.')">.</button>
          <button class="unit-key unit-key-function" @click="showUnitPicker = true">
            <Icon icon="mdi:chevron-down" />
          </button>
        </div>

        <!-- 单位选择弹窗 -->
        <div v-if="showUnitPicker" class="unit-picker-modal" @click="showUnitPicker = false">
          <div class="unit-picker-content" @click.stop>
            <div class="unit-picker-header">
              <span>选择单位</span>
              <button @click="showUnitPicker = false">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="unit-picker-list">
              <button 
                v-for="unit in currentUnits" 
                :key="unit.value"
                class="unit-picker-item"
                :class="{ active: toUnit === unit.value }"
                @click="selectToUnit(unit.value)"
              >
                {{ unit.label }}
                <Icon v-if="toUnit === unit.value" icon="mdi:check" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'

const activeTab = ref('basic')
const input = ref('')
const history = ref('')
const lastResult = ref('')

const clearButtonText = computed(() => {
  if (input.value === '') {
    return 'AC'
  }
  return 'C'
})

const unitType = ref('length')
const fromUnit = ref('m')
const toUnit = ref('km')
const unitValue = ref('0')
const showUnitPicker = ref(false)

const unitTypes = [
  { value: 'length', label: '长度', icon: 'mdi:ruler' },
  { value: 'weight', label: '重量', icon: 'mdi:weight' },
  { value: 'temperature', label: '温度', icon: 'mdi:thermometer' },
  { value: 'volume', label: '体积', icon: 'mdi:cup-water' }
]

const currentUnits = computed(() => units[unitType.value])

const convertedValue = computed(() => {
  if (!unitValue.value || unitValue.value === '0') return '0'
  
  try {
    let result
    const val = parseFloat(unitValue.value) || 0
    
    switch (unitType.value) {
      case 'length':
        result = convertLength(val, fromUnit.value, toUnit.value)
        break
      case 'weight':
        result = convertWeight(val, fromUnit.value, toUnit.value)
        break
      case 'temperature':
        result = convertTemperature(val, fromUnit.value, toUnit.value)
        break
      case 'volume':
        result = convertVolume(val, fromUnit.value, toUnit.value)
        break
      default:
        result = val
    }
    
    // 格式化结果
    if (result < 0.0001 || result > 1000000) {
      return result.toExponential(4)
    }
    return parseFloat(result.toFixed(6)).toString()
  } catch (error) {
    return 'Error'
  }
})

const units = {
  length: [
    { value: 'm', label: '米' },
    { value: 'km', label: '千米' },
    { value: 'cm', label: '厘米' },
    { value: 'mm', label: '毫米' },
    { value: 'in', label: '英寸' },
    { value: 'ft', label: '英尺' },
    { value: 'yd', label: '码' }
  ],
  weight: [
    { value: 'kg', label: '千克' },
    { value: 'g', label: '克' },
    { value: 'mg', label: '毫克' },
    { value: 'lb', label: '磅' },
    { value: 'oz', label: '盎司' }
  ],
  temperature: [
    { value: 'c', label: '摄氏度' },
    { value: 'f', label: '华氏度' },
    { value: 'k', label: '开尔文' }
  ],
  volume: [
    { value: 'l', label: '升' },
    { value: 'ml', label: '毫升' },
    { value: 'gal', label: '加仑' },
    { value: 'qt', label: '夸脱' },
    { value: 'pt', label: '品脱' }
  ]
}

const append = (value) => {
  if (value === '%') {
    handlePercentage()
  } else if (value === '×') {
    input.value += '×'
  } else if (value === '÷') {
    input.value += '÷'
  } else {
    input.value += value
  }
}

const handlePercentage = () => {
  // iOS 风格 % 运算符逻辑
  // 解析当前表达式，找到最后一个运算符
  const operators = ['+', '-', '×', '÷']
  let lastOperatorIndex = -1
  let lastOperator = ''
  
  for (let i = input.value.length - 1; i >= 0; i--) {
    if (operators.includes(input.value[i])) {
      lastOperatorIndex = i
      lastOperator = input.value[i]
      break
    }
  }
  
  if (lastOperatorIndex === -1) {
    // 没有运算符，只是单个数字，直接除以100
    const num = parseFloat(input.value) || 0
    input.value = (num / 100).toString()
  } else {
    // 有运算符，根据运算符类型计算百分比
    const basePart = input.value.substring(0, lastOperatorIndex)
    const percentPart = input.value.substring(lastOperatorIndex + 1)
    
    const baseValue = parseFloat(basePart) || 0
    const percentValue = parseFloat(percentPart) || 0
    
    let result = 0
    switch (lastOperator) {
      case '+':
        // 100 + 15% = 100 + (100 * 0.15) = 115
        result = baseValue + (baseValue * percentValue / 100)
        break
      case '-':
        // 100 - 15% = 100 - (100 * 0.15) = 85
        result = baseValue - (baseValue * percentValue / 100)
        break
      case '×':
        // 100 × 15% = 100 × 0.15 = 15
        result = baseValue * (percentValue / 100)
        break
      case '÷':
        // 100 ÷ 15% = 100 ÷ 0.15 = 666.67
        result = baseValue / (percentValue / 100)
        break
    }
    
    // 格式化结果，避免浮点数精度问题
    input.value = parseFloat(result.toFixed(10)).toString()
  }
}

const backspace = () => {
  if (input.value.length > 0) {
    input.value = input.value.slice(0, -1)
  }
}

const clear = () => {
  if (input.value !== '') {
    input.value = ''
  } else {
    input.value = ''
    history.value = ''
    lastResult.value = ''
  }
}

const toggleSign = () => {
  if (input.value === '') return
  
  const operators = ['+', '-', '×', '÷']
  let lastOperatorIndex = -1
  
  for (let i = input.value.length - 1; i >= 0; i--) {
    if (operators.includes(input.value[i])) {
      lastOperatorIndex = i
      break
    }
  }
  
  if (lastOperatorIndex === -1) {
    if (input.value.startsWith('-')) {
      input.value = input.value.substring(1)
    } else {
      input.value = '-' + input.value
    }
  } else {
    const beforeOperator = input.value.substring(0, lastOperatorIndex + 1)
    const afterOperator = input.value.substring(lastOperatorIndex + 1)
    
    if (afterOperator.startsWith('-')) {
      input.value = beforeOperator + afterOperator.substring(1)
    } else {
      input.value = beforeOperator + '-' + afterOperator
    }
  }
}

const calculate = () => {
  try {
    // 将显示符号转换为 JavaScript 运算符
    const expression = input.value.replace(/×/g, '*').replace(/÷/g, '/')
    const result = new Function('return ' + expression)()
    history.value = `${input.value} = ${result}`
    input.value = parseFloat(result.toFixed(10)).toString()
  } catch (error) {
    input.value = 'Error'
  }
}

const getUnitLabel = (value) => {
  const unit = units[unitType.value].find(u => u.value === value)
  return unit ? unit.label : value
}

const setUnitType = (type) => {
  unitType.value = type
  // 重置为默认单位
  const defaultUnits = {
    length: { from: 'm', to: 'km' },
    weight: { from: 'kg', to: 'g' },
    temperature: { from: 'c', to: 'f' },
    volume: { from: 'l', to: 'ml' }
  }
  fromUnit.value = defaultUnits[type].from
  toUnit.value = defaultUnits[type].to
}

const appendUnitValue = (value) => {
  if (unitValue.value === '0' && value !== '.') {
    unitValue.value = value
  } else if (value === '.' && unitValue.value.includes('.')) {
    return
  } else {
    unitValue.value += value
  }
}

const backspaceUnitValue = () => {
  if (unitValue.value.length > 1) {
    unitValue.value = unitValue.value.slice(0, -1)
  } else {
    unitValue.value = '0'
  }
}

const clearUnitValue = () => {
  unitValue.value = '0'
}

const toggleUnitSign = () => {
  if (unitValue.value === '0') return
  if (unitValue.value.startsWith('-')) {
    unitValue.value = unitValue.value.substring(1)
  } else {
    unitValue.value = '-' + unitValue.value
  }
}

const selectToUnit = (value) => {
  toUnit.value = value
  showUnitPicker.value = false
}

const swapUnits = () => {
  const temp = fromUnit.value
  fromUnit.value = toUnit.value
  toUnit.value = temp
}

const convertLength = (value, from, to) => {
  const meters = {
    m: value,
    km: value * 1000,
    cm: value / 100,
    mm: value / 1000,
    in: value * 0.0254,
    ft: value * 0.3048,
    yd: value * 0.9144
  }[from]
  
  return {
    m: meters,
    km: meters / 1000,
    cm: meters * 100,
    mm: meters * 1000,
    in: meters / 0.0254,
    ft: meters / 0.3048,
    yd: meters / 0.9144
  }[to]
}

const convertWeight = (value, from, to) => {
  const kg = {
    kg: value,
    g: value / 1000,
    mg: value / 1000000,
    lb: value * 0.453592,
    oz: value * 0.0283495
  }[from]
  
  return {
    kg: kg,
    g: kg * 1000,
    mg: kg * 1000000,
    lb: kg / 0.453592,
    oz: kg / 0.0283495
  }[to]
}

const convertTemperature = (value, from, to) => {
  if (from === to) return value
  
  let celsius
  if (from === 'c') celsius = value
  if (from === 'f') celsius = (value - 32) * 5/9
  if (from === 'k') celsius = value - 273.15
  
  if (to === 'c') return celsius
  if (to === 'f') return celsius * 9/5 + 32
  if (to === 'k') return celsius + 273.15
}

const convertVolume = (value, from, to) => {
  const liters = {
    l: value,
    ml: value / 1000,
    gal: value * 3.78541,
    qt: value * 0.946353,
    pt: value * 0.473176
  }[from]
  
  return {
    l: liters,
    ml: liters * 1000,
    gal: liters / 3.78541,
    qt: liters / 0.946353,
    pt: liters / 0.473176
  }[to]
}
</script>

<style scoped>
.calculator-container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.card {
  background: var(--card-bg);
  border-radius: 16px;
  padding: 30px;
  box-shadow: var(--shadow);
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
}

.page-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 24px;
  color: var(--text-primary);
  font-size: 24px;
  font-weight: 600;
}

.page-title svg {
  color: var(--primary-color);
  width: 28px;
  height: 28px;
}

.tabs {
  display: flex;
  margin-bottom: 24px;
  gap: 12px;
}

.tab-btn {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  background: var(--btn-secondary-bg);
  color: var(--text-secondary);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.tab-btn svg {
  width: 14px;
  height: 14px;
}

.tab-btn.active {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  box-shadow: 0 4px 15px rgba(67, 97, 238, 0.3);
}

.calculator {
  max-width: 400px;
  margin: 0 auto;
}

.display {
  background: var(--result-bg);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  text-align: right;
  transition: all 0.3s ease;
  border: 1px solid var(--border-color);
  overflow: hidden;
  word-break: break-all;
}

.history {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 8px;
  min-height: 20px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.input {
  font-size: 36px;
  font-weight: 600;
  color: var(--text-primary);
  min-height: 44px;
  word-break: break-all;
  overflow-wrap: break-word;
  line-height: 1.3;
}

.buttons {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.btn {
  aspect-ratio: 1;
  padding: 0;
  border: none;
  border-radius: 50%;
  font-size: 22px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  width: 100%;
  height: auto;
}

.btn svg {
  width: 22px;
  height: 22px;
}

.btn-number {
  background: #333333;
  color: #ffffff;
}

.btn-number:hover {
  background: #4a4a4a;
}

.btn-number:active {
  background: #555555;
  transform: scale(0.95);
}

.btn-operator {
  background: #ff9f0a;
  color: white;
}

.btn-operator:hover {
  background: #ffb340;
}

.btn-operator:active {
  background: #cc7f08;
  transform: scale(0.95);
}

.btn-function {
  background: #a5a5a5;
  color: #000000;
}

.btn-function:hover {
  background: #c4c4c4;
}

.btn-function:active {
  background: #d4d4d4;
  transform: scale(0.95);
}

.btn-clear {
  background: #a5a5a5;
  color: #000000;
  font-size: 18px;
}

.btn-clear:hover {
  background: #c4c4c4;
}

.btn-clear:active {
  background: #d4d4d4;
  transform: scale(0.95);
}

.btn-equals {
  background: #ff9f0a;
  color: white;
}

.btn-equals:hover {
  background: #ffb340;
}

.btn-equals:active {
  background: #cc7f08;
  transform: scale(0.95);
}

/* 单位转换器 - iOS风格 */
.unit-converter {
  margin-top: 20px;
}

/* 分段控制器 */
.unit-type-segmented {
  display: flex;
  background: #e5e5ea;
  border-radius: 10px;
  padding: 4px;
  margin-bottom: 24px;
  gap: 4px;
}

.segment-btn {
  flex: 1;
  padding: 10px 8px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #000;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.segment-btn svg {
  width: 16px;
  height: 16px;
}

.segment-btn.active {
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

/* 输入显示区域 */
.unit-input-section {
  background: #1c1c1e;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
}

.unit-display {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 12px 0;
  cursor: pointer;
  border-bottom: 1px solid #3a3a3c;
}

.unit-display:last-of-type {
  border-bottom: none;
}

.unit-value {
  font-size: 36px;
  font-weight: 300;
  color: #fff;
  font-variant-numeric: tabular-nums;
}

.unit-name {
  font-size: 16px;
  color: #ff9f0a;
  font-weight: 500;
}

.unit-swap-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: #3a3a3c;
  color: #ff9f0a;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 8px auto;
  transition: all 0.2s ease;
}

.unit-swap-btn:hover {
  background: #48484a;
}

.unit-swap-btn:active {
  transform: scale(0.95);
}

.unit-swap-btn svg {
  width: 20px;
  height: 20px;
}

/* 数字键盘 */
.unit-keyboard {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.unit-key {
  aspect-ratio: 1;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: #333333;
  color: #fff;
  font-size: 24px;
  font-weight: 400;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.unit-key:hover {
  background: #4a4a4a;
}

.unit-key:active {
  background: #555555;
  transform: scale(0.95);
}

.unit-key-function {
  background: #a5a5a5;
  color: #000;
}

.unit-key-function:hover {
  background: #c4c4c4;
}

.unit-key-function:active {
  background: #d4d4d4;
}

.unit-key-zero {
  grid-column: span 1;
}

.unit-key svg {
  width: 24px;
  height: 24px;
}

/* 单位选择弹窗 */
.unit-picker-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.unit-picker-content {
  background: #1c1c1e;
  border-radius: 16px 16px 0 0;
  width: 100%;
  max-width: 400px;
  max-height: 70vh;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

.unit-picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #3a3a3c;
}

.unit-picker-header span {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.unit-picker-header button {
  background: none;
  border: none;
  color: #ff9f0a;
  cursor: pointer;
  padding: 4px;
}

.unit-picker-header button svg {
  width: 24px;
  height: 24px;
}

.unit-picker-list {
  max-height: 50vh;
  overflow-y: auto;
  padding: 8px 0;
}

.unit-picker-item {
  width: 100%;
  padding: 16px 20px;
  border: none;
  background: transparent;
  color: #fff;
  font-size: 17px;
  text-align: left;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: background 0.15s ease;
}

.unit-picker-item:hover {
  background: #2c2c2e;
}

.unit-picker-item.active {
  color: #ff9f0a;
}

.unit-picker-item svg {
  width: 20px;
  height: 20px;
}

.result-main {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 20px;
  color: var(--text-primary);
}

.result-main svg {
  color: var(--success-color);
  width: 20px;
  height: 20px;
}

.result-main strong {
  color: var(--primary-color);
  font-size: 24px;
}

@media (max-width: 480px) {
  .calculator-container {
    padding: 12px;
  }

  .card {
    padding: 20px;
    border-radius: 12px;
  }

  .page-title {
    font-size: 20px;
    gap: 8px;
  }

  .page-title svg {
    width: 22px;
    height: 22px;
  }

  .tabs {
    gap: 8px;
  }

  .tab-btn {
    padding: 10px 12px;
    font-size: 14px;
  }

  .tab-btn span {
    display: none;
  }

  .tab-btn svg {
    width: 18px;
    height: 18px;
  }

  .buttons {
    gap: 10px;
  }

  .btn {
    font-size: 20px;
  }

  .btn svg {
    width: 20px;
    height: 20px;
  }

  .btn-clear {
    font-size: 16px;
  }

  .display {
    padding: 15px;
  }

  .history {
    font-size: 12px;
  }

  .input {
    font-size: 32px;
  }

  /* 单位转换器移动端适配 */
  .unit-type-segmented {
    margin-bottom: 16px;
  }

  .segment-btn {
    padding: 8px 4px;
    font-size: 12px;
  }

  .segment-btn svg {
    width: 14px;
    height: 14px;
  }

  .unit-input-section {
    padding: 16px;
    margin-bottom: 16px;
  }

  .unit-value {
    font-size: 28px;
  }

  .unit-name {
    font-size: 14px;
  }

  .unit-keyboard {
    gap: 8px;
  }

  .unit-key {
    font-size: 20px;
  }

  .unit-key svg {
    width: 20px;
    height: 20px;
  }

  .unit-row {
    flex-direction: column;
    gap: 12px;
  }

  .unit-col {
    width: 100%;
  }

  .swap-btn {
    width: 100%;
    height: 36px;
  }
}

@media (max-width: 360px) {
  .display {
    padding: 12px;
  }

  .history {
    font-size: 11px;
  }

  .input {
    font-size: 20px;
  }

  .buttons {
    gap: 6px;
  }

  .btn {
    padding: 12px;
    font-size: 16px;
  }

  .btn svg {
    width: 14px;
    height: 14px;
  }
}
</style>
