type Series = { data: Array<{ x: number; y: number }> };

function a(series: Series, substract: Series[]): void {
  for (let x = 0; x < series.data.length; x++) {
    let newValue = series.data[x]!.y;
    for (const otherSeries of substract) {
      newValue -= otherSeries.data[x]!.y; // x is used to index other arrays
    }
    series.data[x]!.y = newValue;
  }
}

// This should trigger the rule
const array = [1, 2, 3];
for (let i = 0; i < array.length; i++) {
  console.log(array[i]);
}