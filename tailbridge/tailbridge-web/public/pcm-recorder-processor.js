class PCMProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
  }

  process(inputs, outputs, parameters) {
    // We expect a single input with a single channel
    if (inputs.length > 0 && inputs[0].length > 0) {
      const inputChannel = inputs[0][0];
      // Create a copy of the input data to send over the port
      const inputCopy = new Float32Array(inputChannel);
      this.port.postMessage(inputCopy);
    }
    return true; // Keep the processor alive
  }
}

registerProcessor("pcm-recorder-processor", PCMProcessor);
