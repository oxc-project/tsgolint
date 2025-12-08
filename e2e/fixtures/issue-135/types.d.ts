declare module 'node2:test' {
    import TestFn = test.TestFn;
    type TestOptions = {};
    function test(name?: string, fn?: TestFn): Promise<void>;
    function test(name?: string, options?: TestOptions, fn?: TestFn): Promise<void>;
    function test(options?: TestOptions, fn?: TestFn): Promise<void>;
    function test(fn?: TestFn): Promise<void>;
    namespace test {
        export { test };
        export { suite as describe, test as it };
    }
    namespace test {
        function run(options?: {}): {};
        function suite(name?: string, options?: TestOptions, fn?: SuiteFn): Promise<void>;
        function suite(name?: string, fn?: SuiteFn): Promise<void>;
        function suite(options?: TestOptions, fn?: SuiteFn): Promise<void>;
        function suite(fn?: SuiteFn): Promise<void>;
        namespace suite {
            function skip(name?: string, options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function skip(name?: string, fn?: SuiteFn): Promise<void>;
            function skip(options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function skip(fn?: SuiteFn): Promise<void>;
            function todo(name?: string, options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function todo(name?: string, fn?: SuiteFn): Promise<void>;
            function todo(options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function todo(fn?: SuiteFn): Promise<void>;
            function only(name?: string, options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function only(name?: string, fn?: SuiteFn): Promise<void>;
            function only(options?: TestOptions, fn?: SuiteFn): Promise<void>;
            function only(fn?: SuiteFn): Promise<void>;
        }
        type TestFn = (t: {}, done: (result?: any) => void) => void | Promise<void>;
        type SuiteFn = (s: {}) => void | Promise<void>;
    }
    export = test;
}
