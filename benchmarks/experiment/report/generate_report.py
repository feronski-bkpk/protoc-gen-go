#!/usr/bin/env python3
"""
Генератор графиков и авто-отчёта BENCHMARK_REPORT.md.
Использование:
  python3 generate_report.py              # Графики + отчёт (на основе имеющихся результатов)
  python3 generate_report.py --charts-only  # Только графики
  python3 generate_report.py --report-only  # Только Markdown-отчёт
Требования: matplotlib, numpy
"""

import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import numpy as np
import os
import re
import sys
from datetime import datetime

# ============================================================
# ПАРСИНГ РЕЗУЛЬТАТОВ БЕНЧМАРКОВ
# ============================================================

def parse_benchmark_results(filepath):
    """
    Парсит результаты go test -bench и python timeit.
    Возвращает словарь {benchmark_name: {'ns': float, 'mem': int, 'allocs': int}}
    """
    results = {}
    if not os.path.exists(filepath):
        print(f"  ⚠  Файл не найден: {filepath}")
        return results

    with open(filepath, 'r') as f:
        for line in f:
            # Go: BenchmarkBuild_64B_0ext-4   7669753   280.0 ns/op   304 B/op   4 allocs/op
            match = re.match(
                r'(Benchmark\S+)\s+(\d+)\s+([\d.]+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op',
                line
            )
            if match:
                name = re.sub(r'-\d+$', '', match.group(1))
                results[name] = {
                    'ns': float(match.group(3)),
                    'mem': int(match.group(4)),
                    'allocs': int(match.group(5)),
                }
                continue

            # Python Construct: ConstructBuild_64B_0ext: 64339.5 ns/op
            match = re.match(r'(Construct\S+):\s+([\d.]+)\s+ns/op', line)
            if match:
                results[match.group(1)] = {
                    'ns': float(match.group(2)),
                    'mem': 0,
                    'allocs': 0,
                }

    return results


def load_all_data():
    """Загружает все доступные результаты бенчмарков."""
    report_dir = os.path.dirname(os.path.abspath(__file__))
    experiment_dir = os.path.dirname(report_dir)

    data = {
        'dsl':  parse_benchmark_results(os.path.join(experiment_dir, 'results_dsl.txt')),
        'hand': parse_benchmark_results(os.path.join(experiment_dir, 'results_hand.txt')),
        'pb':   parse_benchmark_results(os.path.join(experiment_dir, 'results_pb.txt')),
        'construct': parse_benchmark_results(os.path.join(experiment_dir, 'results_construct.txt')),
    }

    # Определяем, какие инструменты доступны
    available = [k for k, v in data.items() if v]
    print(f"  ✓ Загружены данные: {', '.join(available)}")
    return data


# ============================================================
# ГЛОБАЛЬНЫЕ НАСТРОЙКИ
# ============================================================

report_dir = os.path.dirname(os.path.abspath(__file__))
experiment_dir = os.path.dirname(report_dir)

COLORS = {
    'dsl':       '#2196F3',
    'hand':      '#4CAF50',
    'pb':        '#FF9800',
    'construct': '#F44336',
}

STYLES = {
    'dsl':       {'marker': 'o', 'linestyle': '-',  'linewidth': 2.5, 'markersize': 9, 'zorder': 5},
    'hand':      {'marker': 's', 'linestyle': '--', 'linewidth': 2.0, 'markersize': 8, 'zorder': 4},
    'pb':        {'marker': '^', 'linestyle': ':',  'linewidth': 2.5, 'markersize': 9, 'zorder': 3},
    'construct': {'marker': 'D', 'linestyle': '-.', 'linewidth': 2.0, 'markersize': 8, 'zorder': 2},
}

LABELS = {
    'dsl':       'protoc-gen-go (DSL) — данный инструмент',
    'hand':      'Ручная реализация Go — идеализированный baseline',
    'pb':        'Google Protobuf — индустриальный стандарт',
    'construct': 'Python Construct — декларативная библиотека',
}

plt.rcParams.update({
    'figure.figsize': (12, 7),
    'font.size': 12,
    'axes.titlesize': 14,
    'axes.labelsize': 12,
    'legend.fontsize': 11,
    'axes.grid': True,
    'grid.alpha': 0.3,
    'grid.linestyle': '--',
})

SIZE_KEYS = ['64B_0ext', '256B_2ext', '1024B_4ext', '4096B_4ext', '64K_4ext']
SIZE_LABELS = ['64 B', '256 B', '1 KB', '4 KB', '64 KB']
EXT_KEYS = ['1024B_Ext0', '1024B_Ext1', '1024B_Ext2', '1024B_Ext3', '1024B_Ext4']

BENCH_PREFIXES = {
    'dsl':  {'build': 'BenchmarkBuild_',     'parse': 'BenchmarkParse_'},
    'hand': {'build': 'BenchmarkHandBuild_', 'parse': 'BenchmarkHandParse_'},
    'pb':   {'build': 'BenchmarkPBBuild_',   'parse': 'BenchmarkPBParse_'},
    'construct': {'build': 'ConstructBuild_', 'parse': 'ConstructParse_'},
}


def get_val(data, tool, prefix, key, field='ns'):
    """Безопасно извлекает значение из данных бенчмарка."""
    if tool not in data or not data[tool]:
        return None
    name = BENCH_PREFIXES[tool][prefix] + key
    entry = data[tool].get(name, {})
    return entry.get(field, None)


def save_and_close(fig, filename):
    filepath = os.path.join(report_dir, filename)
    fig.tight_layout(pad=1.5)
    fig.savefig(filepath, dpi=150, bbox_inches='tight', facecolor='white')
    plt.close(fig)
    print(f"  ✓ {filename}")


def plot_line(ax, x, y, tool_key):
    if y is None or all(v is None for v in y):
        return
    style = STYLES[tool_key]
    valid_x = [xi for xi, yi in zip(x, y) if yi is not None]
    valid_y = [yi for yi in y if yi is not None]
    if valid_y:
        ax.plot(valid_x, valid_y,
                marker=style['marker'], linestyle=style['linestyle'],
                linewidth=style['linewidth'], markersize=style['markersize'],
                color=COLORS[tool_key], label=LABELS[tool_key], zorder=style['zorder'])


# ============================================================
# ГЕНЕРАЦИЯ ГРАФИКОВ
# ============================================================

def generate_all_charts(data):
    """Генерирует все графики на основе загруженных данных."""

    # --- График 1: Marshal vs Размер ---
    fig, ax = plt.subplots(figsize=(14, 7))
    plot_line(ax, SIZE_LABELS, [get_val(data, 'dsl', 'build', k) for k in SIZE_KEYS], 'dsl')
    plot_line(ax, SIZE_LABELS, [get_val(data, 'hand', 'build', k) for k in SIZE_KEYS], 'hand')
    plot_line(ax, SIZE_LABELS, [get_val(data, 'pb', 'build', k) for k in SIZE_KEYS], 'pb')

    if data.get('construct'):
        ax_c = ax.twinx()
        cx = SIZE_LABELS[:4]
        cy = [get_val(data, 'construct', 'build', k) for k in SIZE_KEYS[:4]]
        if any(v is not None for v in cy):
            ax_c.plot(cx, cy, marker=STYLES['construct']['marker'],
                       linestyle=STYLES['construct']['linestyle'],
                       linewidth=STYLES['construct']['linewidth'],
                       markersize=STYLES['construct']['markersize'],
                       color=COLORS['construct'], label=LABELS['construct'],
                       zorder=STYLES['construct']['zorder'])
            ax_c.set_yscale('log')
            ax_c.set_ylabel('Время Marshal — Python Construct (ns/op)', color=COLORS['construct'])
            lines1, labels1 = ax.get_legend_handles_labels()
            lines2, labels2 = ax_c.get_legend_handles_labels()
            ax.legend(lines1 + lines2, labels1 + labels2, loc='upper left', framealpha=0.9)
        else:
            ax.legend(loc='upper left', framealpha=0.9)
    else:
        ax.legend(loc='upper left', framealpha=0.9)

    ax.set_yscale('log')
    ax.set_xlabel('Размер полезной нагрузки')
    ax.set_ylabel('Время Marshal — Go (ns/op)', color='black')
    ax.set_title('Эксперимент А: Время сборки (Marshal) vs Размер полезной нагрузки')
    ax.set_ylim(bottom=50)
    save_and_close(fig, 'chart_a_marshal.png')

    # --- График 2: Unmarshal vs Размер ---
    fig, ax = plt.subplots(figsize=(14, 7))
    plot_line(ax, SIZE_LABELS, [get_val(data, 'dsl', 'parse', k) for k in SIZE_KEYS], 'dsl')
    plot_line(ax, SIZE_LABELS, [get_val(data, 'hand', 'parse', k) for k in SIZE_KEYS], 'hand')
    plot_line(ax, SIZE_LABELS, [get_val(data, 'pb', 'parse', k) for k in SIZE_KEYS], 'pb')

    if data.get('construct'):
        ax_c = ax.twinx()
        cx = SIZE_LABELS[:4]
        cy = [get_val(data, 'construct', 'parse', k) for k in SIZE_KEYS[:4]]
        if any(v is not None for v in cy):
            ax_c.plot(cx, cy, marker=STYLES['construct']['marker'],
                       linestyle=STYLES['construct']['linestyle'],
                       linewidth=STYLES['construct']['linewidth'],
                       markersize=STYLES['construct']['markersize'],
                       color=COLORS['construct'], label=LABELS['construct'],
                       zorder=STYLES['construct']['zorder'])
            ax_c.set_yscale('log')
            ax_c.set_ylabel('Время Unmarshal — Python Construct (ns/op)', color=COLORS['construct'])
            lines1, labels1 = ax.get_legend_handles_labels()
            lines2, labels2 = ax_c.get_legend_handles_labels()
            ax.legend(lines1 + lines2, labels1 + labels2, loc='upper left', framealpha=0.9)
        else:
            ax.legend(loc='upper left', framealpha=0.9)
    else:
        ax.legend(loc='upper left', framealpha=0.9)

    ax.set_yscale('log')
    ax.set_xlabel('Размер полезной нагрузки')
    ax.set_ylabel('Время Unmarshal — Go (ns/op)', color='black')
    ax.set_title('Эксперимент А: Время разбора (Unmarshal) vs Размер полезной нагрузки')
    ax.set_ylim(bottom=20)
    save_and_close(fig, 'chart_a_unmarshal.png')

    # --- График 3: Throughput ---
    fig, ax = plt.subplots(figsize=(12, 6))
    pkt_sizes = [106, 298, 1098, 4170, 65536]

    def tp(values, sizes):
        if not values:
            return None
        return [(s / (v * 1e-9) / 1e6) if v else 0 for s, v in zip(sizes, values)]

    plot_line(ax, SIZE_LABELS, tp([get_val(data, 'dsl', 'build', k) for k in SIZE_KEYS], pkt_sizes), 'dsl')
    plot_line(ax, SIZE_LABELS, tp([get_val(data, 'hand', 'build', k) for k in SIZE_KEYS], pkt_sizes), 'hand')
    plot_line(ax, SIZE_LABELS, tp([get_val(data, 'pb', 'build', k) for k in SIZE_KEYS],
                                   [124, 332, 1180, 4350, 67000]), 'pb')
    ax.set_xlabel('Размер полезной нагрузки')
    ax.set_ylabel('Пропускная способность (MB/s)')
    ax.set_title('Эксперимент А: Пропускная способность при сборке (Marshal)')
    ax.legend(loc='upper left', framealpha=0.9)
    save_and_close(fig, 'chart_a_throughput.png')

    # --- График 4: Аллокации при Marshal ---
    fig, ax = plt.subplots(figsize=(12, 6))
    x = np.arange(len(SIZE_LABELS))
    width = 0.25

    dsl_mem = [get_val(data, 'dsl', 'build', k, 'mem') or 0 for k in SIZE_KEYS]
    hand_mem = [get_val(data, 'hand', 'build', k, 'mem') or 0 for k in SIZE_KEYS]
    pb_mem = [get_val(data, 'pb', 'build', k, 'mem') or 0 for k in SIZE_KEYS]

    bars1 = ax.bar(x - width, dsl_mem, width, color=COLORS['dsl'], label=LABELS['dsl'], zorder=3)
    bars2 = ax.bar(x, hand_mem, width, color=COLORS['hand'], label=LABELS['hand'], zorder=3)
    bars3 = ax.bar(x + width, pb_mem, width, color=COLORS['pb'], label=LABELS['pb'], zorder=3)

    for bars in [bars1, bars2, bars3]:
        for bar in bars:
            h = bar.get_height()
            label = f'{h/1024:.0f}K' if h > 1000 else f'{h:.0f}'
            ax.annotate(label, xy=(bar.get_x() + bar.get_width()/2, h),
                         xytext=(0, 3), textcoords="offset points",
                         ha='center', va='bottom', fontsize=7)

    ax.set_xticks(x)
    ax.set_xticklabels(SIZE_LABELS)
    ax.set_xlabel('Размер полезной нагрузки')
    ax.set_ylabel('Выделено памяти (B/op)')
    ax.set_title('Эксперимент А: Аллокации памяти при сборке (Marshal)')
    ax.legend(loc='upper left', framealpha=0.9)
    ax.set_yscale('log')
    save_and_close(fig, 'chart_a_allocs.png')

    # --- График 5: Marshal vs Расширения ---
    ext_counts = [0, 1, 2, 3, 4]
    fig, ax = plt.subplots(figsize=(12, 6))
    plot_line(ax, ext_counts, [get_val(data, 'dsl', 'build', k) for k in EXT_KEYS], 'dsl')
    plot_line(ax, ext_counts, [get_val(data, 'hand', 'build', k) for k in EXT_KEYS], 'hand')
    plot_line(ax, ext_counts, [get_val(data, 'pb', 'build', k) for k in EXT_KEYS], 'pb')
    ax.set_xlabel('Количество расширений IPv6')
    ax.set_ylabel('Время Marshal (ns/op)')
    ax.set_title('Эксперимент Б: Время сборки vs Количество расширений (payload = 1024 B)')
    ax.legend(loc='upper left', framealpha=0.9)
    ax.set_xticks(ext_counts)
    save_and_close(fig, 'chart_b_marshal.png')

    # --- График 6: Unmarshal vs Расширения ---
    fig, ax = plt.subplots(figsize=(12, 6))
    plot_line(ax, ext_counts, [get_val(data, 'dsl', 'parse', k) for k in EXT_KEYS], 'dsl')
    plot_line(ax, ext_counts, [get_val(data, 'hand', 'parse', k) for k in EXT_KEYS], 'hand')
    plot_line(ax, ext_counts, [get_val(data, 'pb', 'parse', k) for k in EXT_KEYS], 'pb')
    ax.set_xlabel('Количество расширений IPv6')
    ax.set_ylabel('Время Unmarshal (ns/op)')
    ax.set_title('Эксперимент Б: Время разбора vs Количество расширений (payload = 1024 B)')
    ax.legend(loc='upper left', framealpha=0.9)
    ax.set_xticks(ext_counts)
    save_and_close(fig, 'chart_b_unmarshal.png')

    # --- График 7: Аллокации при Marshal (эксперимент Б) ---
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 6))
    x = np.arange(len(ext_counts))
    width = 0.25

    dsl_allocs = [get_val(data, 'dsl', 'build', k, 'allocs') or 0 for k in EXT_KEYS]
    hand_allocs = [get_val(data, 'hand', 'build', k, 'allocs') or 0 for k in EXT_KEYS]
    pb_allocs = [get_val(data, 'pb', 'build', k, 'allocs') or 0 for k in EXT_KEYS]
    dsl_amem = [get_val(data, 'dsl', 'build', k, 'mem') or 0 for k in EXT_KEYS]
    hand_amem = [get_val(data, 'hand', 'build', k, 'mem') or 0 for k in EXT_KEYS]
    pb_amem = [get_val(data, 'pb', 'build', k, 'mem') or 0 for k in EXT_KEYS]

    ax1.bar(x - width, dsl_allocs, width, color=COLORS['dsl'], label=LABELS['dsl'], zorder=3)
    ax1.bar(x, hand_allocs, width, color=COLORS['hand'], label=LABELS['hand'], zorder=3)
    ax1.bar(x + width, pb_allocs, width, color=COLORS['pb'], label=LABELS['pb'], zorder=3)
    ax1.set_xticks(x)
    ax1.set_xticklabels(ext_counts)
    ax1.set_xlabel('Количество расширений')
    ax1.set_ylabel('Количество аллокаций')
    ax1.set_title('Количество аллокаций')
    ax1.legend(loc='upper left', framealpha=0.9)

    ax2.bar(x - width, dsl_amem, width, color=COLORS['dsl'], label=LABELS['dsl'], zorder=3)
    ax2.bar(x, hand_amem, width, color=COLORS['hand'], label=LABELS['hand'], zorder=3)
    ax2.bar(x + width, pb_amem, width, color=COLORS['pb'], label=LABELS['pb'], zorder=3)
    ax2.set_xticks(x)
    ax2.set_xticklabels(ext_counts)
    ax2.set_xlabel('Количество расширений')
    ax2.set_ylabel('Выделено памяти (B/op)')
    ax2.set_title('Объём выделенной памяти')
    ax2.legend(loc='upper left', framealpha=0.9)

    fig.suptitle('Эксперимент Б: Аллокации при сборке vs Количество расширений (payload = 1024 B)',
                 fontsize=14, y=1.02)
    save_and_close(fig, 'chart_b_allocs.png')

    # --- График 8: Сравнение с Construct ---
    if data.get('construct'):
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 7))
        x = np.arange(4)
        width = 0.3

        dsl_b = [get_val(data, 'dsl', 'build', k) for k in SIZE_KEYS[:4]]
        cons_b = [get_val(data, 'construct', 'build', k) for k in SIZE_KEYS[:4]]
        bar1 = ax1.bar(x - width/2, dsl_b, width, color=COLORS['dsl'], label=LABELS['dsl'], zorder=3)
        ax1_t = ax1.twinx()
        bar2 = ax1_t.bar(x + width/2, cons_b, width, color=COLORS['construct'], alpha=0.6,
                          label=LABELS['construct'], zorder=2)
        ax1.set_xticks(x)
        ax1.set_xticklabels(SIZE_LABELS[:4])
        ax1.set_ylabel('Go (ns/op)', color=COLORS['dsl'])
        ax1_t.set_ylabel('Python (ns/op)', color=COLORS['construct'])
        ax1.set_title('Marshal')
        lines1, labels1 = ax1.get_legend_handles_labels()
        lines2, labels2 = ax1_t.get_legend_handles_labels()
        ax1.legend(lines1 + lines2, labels1 + labels2, loc='upper left', fontsize=9)

        dsl_p = [get_val(data, 'dsl', 'parse', k) for k in SIZE_KEYS[:4]]
        cons_p = [get_val(data, 'construct', 'parse', k) for k in SIZE_KEYS[:4]]
        bar3 = ax2.bar(x - width/2, dsl_p, width, color=COLORS['dsl'], label=LABELS['dsl'], zorder=3)
        ax2_t = ax2.twinx()
        bar4 = ax2_t.bar(x + width/2, cons_p, width, color=COLORS['construct'], alpha=0.6,
                          label=LABELS['construct'], zorder=2)
        ax2.set_xticks(x)
        ax2.set_xticklabels(SIZE_LABELS[:4])
        ax2.set_ylabel('Go (ns/op)', color=COLORS['dsl'])
        ax2_t.set_ylabel('Python (ns/op)', color=COLORS['construct'])
        ax2.set_title('Unmarshal')
        lines1, labels1 = ax2.get_legend_handles_labels()
        lines2, labels2 = ax2_t.get_legend_handles_labels()
        ax2.legend(lines1 + lines2, labels1 + labels2, loc='upper left', fontsize=9)

        fig.suptitle('Сравнение: protoc-gen-go (Go) vs Python Construct\n(разные шкалы из-за разницы в 100–750×)',
                     fontsize=14, y=1.02)
        save_and_close(fig, 'chart_construct.png')

    # --- График 9: Размер схемы и кода ---
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 6))
    schema_tools = ['DSL', 'Protobuf']
    schema_bytes = [1210, 560]
    schema_colors = [COLORS['dsl'], COLORS['pb']]
    bars = ax1.bar(schema_tools, schema_bytes, color=schema_colors, width=0.4)
    for bar, val in zip(bars, schema_bytes):
        ax1.annotate(f'{val} B', (bar.get_x() + bar.get_width()/2, bar.get_height()),
                      textcoords="offset points", xytext=(0, 5), ha='center', fontsize=12, fontweight='bold')
    ax1.set_ylabel('Байт')
    ax1.set_title('Размер схемы (DSL vs .proto)')

    code_tools = ['DSL\n(сгенер.)', 'Ручной Go', 'Protobuf\n(сгенер.)', 'Construct\n(Python)']
    code_lines = [609, 270, 350, 166]
    code_colors = [COLORS['dsl'], COLORS['hand'], COLORS['pb'], COLORS['construct']]
    bars = ax2.bar(code_tools, code_lines, color=code_colors, width=0.5)
    for bar, val in zip(bars, code_lines):
        ax2.annotate(f'{val} строк', (bar.get_x() + bar.get_width()/2, bar.get_height()),
                      textcoords="offset points", xytext=(0, 5), ha='center', fontsize=12, fontweight='bold')
    ax2.set_ylabel('Строк кода')
    ax2.set_title('Размер кода')

    fig.suptitle('Эксперимент В: Размер схемы и сгенерированного/написанного кода', fontsize=14, y=1.02)
    save_and_close(fig, 'chart_c_size.png')


# ============================================================
# ГЕНЕРАЦИЯ MARKDOWN-ОТЧЁТА
# ============================================================

def generate_markdown_report(data):
    """Генерирует BENCHMARK_REPORT_<timestamp>.md с таблицами из реальных данных."""
    now = datetime.now()
    now_str = now.strftime('%Y-%m-%d %H:%M')
    timestamp = now.strftime('%Y%m%d_%H%M%S')

    def ns(key, tool='dsl', prefix='build'):
        v = get_val(data, tool, prefix, key)
        return v if v else None

    def fmt_ns(key, tool='dsl', prefix='build'):
        v = ns(key, tool, prefix)
        return f'{v:.0f}' if v else '—'

    def compare(dsl_key, hand_key, prefix='build'):
        d = ns(dsl_key, 'dsl', prefix)
        h = ns(hand_key, 'hand', prefix)
        if d and h and h > 0:
            diff = (d / h - 1) * 100
            sign = '+' if diff > 0 else ''
            return f'{sign}{diff:.1f}%'
        return '—'

    has_hand = bool(data.get('hand'))
    has_pb = bool(data.get('pb'))
    has_construct = bool(data.get('construct'))

    report = f"""# Отчёт о бенчмарках protoc-gen-go

> **Сгенерирован автоматически:** {now_str}
> **Платформа:** {os.uname().sysname} {os.uname().machine}, Python {sys.version.split()[0]}
> **Доступные данные:** DSL{", Hand" if has_hand else ""}{", Protobuf" if has_pb else ""}{", Construct" if has_construct else ""}

---

## Эксперимент А: Варьирование размера полезной нагрузки

### Marshal (сборка)

| Размер | DSL (ns/op){' | Hand (ns/op) | Разница' if has_hand else ''}{' | Protobuf (ns/op) | DSL vs PB' if has_pb else ''} |
|:------:|:----------:{'|:----------:|:-------:' if has_hand else ''}{'|:----------------:|:---------:' if has_pb else ''} |
"""

    for i, key in enumerate(SIZE_KEYS):
        row = f"| {SIZE_LABELS[i]} | {fmt_ns(key, 'dsl', 'build')} |"
        if has_hand:
            row += f" {fmt_ns(key, 'hand', 'build')} | {compare(key, key, 'build')} |"
        if has_pb:
            row += f" {fmt_ns(key, 'pb', 'build')} |"
            d_val = ns(key, 'dsl', 'build')
            p_val = ns(key, 'pb', 'build')
            if d_val and p_val and p_val > 0:
                row += f" **×{p_val/d_val:.1f}** |"
            else:
                row += " — |"
        report += row + "\n"

    report += f"""
![Marshal vs Размер](chart_a_marshal.png)

### Unmarshal (разбор)

| Размер | DSL (ns/op){' | Hand (ns/op) | Разница' if has_hand else ''}{' | Protobuf (ns/op) | DSL vs PB' if has_pb else ''} |
|:------:|:----------:{'|:----------:|:-------:' if has_hand else ''}{'|:----------------:|:---------:' if has_pb else ''} |
"""

    for i, key in enumerate(SIZE_KEYS):
        row = f"| {SIZE_LABELS[i]} | {fmt_ns(key, 'dsl', 'parse')} |"
        if has_hand:
            row += f" {fmt_ns(key, 'hand', 'parse')} | {compare(key, key, 'parse')} |"
        if has_pb:
            row += f" {fmt_ns(key, 'pb', 'parse')} |"
            d_val = ns(key, 'dsl', 'parse')
            p_val = ns(key, 'pb', 'parse')
            if d_val and p_val and p_val > 0:
                row += f" **×{p_val/d_val:.1f}** |"
            else:
                row += " — |"
        report += row + "\n"

    report += f"""
![Unmarshal vs Размер](chart_a_unmarshal.png)

### Пропускная способность (Throughput)

![Throughput](chart_a_throughput.png)

### Аллокации памяти

![Аллокации](chart_a_allocs.png)

---

## Эксперимент Б: Варьирование количества расширений (payload = 1024 B)

### Marshal

| Расш. | DSL (ns/op){' | Hand (ns/op) | Разница' if has_hand else ''}{' | Protobuf (ns/op) | DSL vs PB' if has_pb else ''} |
|:-----:|:----------:{'|:----------:|:-------:' if has_hand else ''}{'|:----------------:|:---------:' if has_pb else ''} |
"""

    for i, key in enumerate(EXT_KEYS):
        row = f"| {i} | {fmt_ns(key, 'dsl', 'build')} |"
        if has_hand:
            row += f" {fmt_ns(key, 'hand', 'build')} | {compare(key, key, 'build')} |"
        if has_pb:
            row += f" {fmt_ns(key, 'pb', 'build')} |"
            d_val = ns(key, 'dsl', 'build')
            p_val = ns(key, 'pb', 'build')
            if d_val and p_val and p_val > 0:
                row += f" **×{p_val/d_val:.1f}** |"
            else:
                row += " — |"
        report += row + "\n"

    report += f"""
![Marshal vs Расширения](chart_b_marshal.png)

### Unmarshal

| Расш. | DSL (ns/op){' | Hand (ns/op) | Разница' if has_hand else ''}{' | Protobuf (ns/op) | DSL vs PB' if has_pb else ''} |
|:-----:|:----------:{'|:----------:|:-------:' if has_hand else ''}{'|:----------------:|:---------:' if has_pb else ''} |
"""

    for i, key in enumerate(EXT_KEYS):
        row = f"| {i} | {fmt_ns(key, 'dsl', 'parse')} |"
        if has_hand:
            row += f" {fmt_ns(key, 'hand', 'parse')} | {compare(key, key, 'parse')} |"
        if has_pb:
            row += f" {fmt_ns(key, 'pb', 'parse')} |"
            d_val = ns(key, 'dsl', 'parse')
            p_val = ns(key, 'pb', 'parse')
            if d_val and p_val and p_val > 0:
                row += f" **×{p_val/d_val:.1f}** |"
            else:
                row += " — |"
        report += row + "\n"

    report += f"""
![Unmarshal vs Расширения](chart_b_unmarshal.png)

### Аллокации

![Аллокации vs Расширения](chart_b_allocs.png)

---

## Размер схемы и кода

![Размер схемы и кода](chart_c_size.png)

---

*Отчёт сгенерирован автоматически: {now_str}*
"""

    # Сохраняем с временной меткой
    filename = f'BENCHMARK_REPORT_{timestamp}.md'
    filepath = os.path.join(report_dir, filename)
    with open(filepath, 'w') as f:
        f.write(report)
    print(f"  ✓ Отчёт сохранён: {filename}")

    # Создаём символьную ссылку на последний отчёт
    latest_path = os.path.join(report_dir, 'BENCHMARK_REPORT_LATEST.md')
    if os.path.exists(latest_path) or os.path.islink(latest_path):
        os.remove(latest_path)
    os.symlink(filename, latest_path)
    print(f"  ✓ Ссылка на последний: BENCHMARK_REPORT_LATEST.md -> {filename}")


# ============================================================
# MAIN
# ============================================================

if __name__ == "__main__":
    print("=" * 60)
    print("  ГЕНЕРАТОР ОТЧЁТА PROTOCOL-GEN-GO")
    print("=" * 60)
    print()

    charts_only = '--charts-only' in sys.argv
    report_only = '--report-only' in sys.argv

    # Загружаем данные
    print("Загрузка результатов бенчмарков...")
    data = load_all_data()

    if not data.get('dsl'):
        print("Нет данных DSL. Запустите make bench-report или make experiment.")
        sys.exit(1)

    # Графики
    if not report_only:
        print()
        print("Генерация графиков...")
        generate_all_charts(data)
        print()
        print("Сгенерированные файлы:")
        for f in sorted(os.listdir(report_dir)):
            if f.endswith('.png'):
                size_kb = os.path.getsize(os.path.join(report_dir, f)) / 1024
                print(f"  📊 {f} ({size_kb:.0f} KB)")

    # Markdown-отчёт
    if not charts_only:
        print()
        print("Генерация Markdown-отчёта...")
        generate_markdown_report(data)

    print()
    print("=" * 60)
    print("  ГОТОВО!")
    print("=" * 60)
