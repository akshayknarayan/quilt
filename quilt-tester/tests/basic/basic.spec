(import "github.com/NetSys/quilt/quilt-tester/config/infrastructure")

// Using unique Namespaces will allow multiple Quilt instances to run on the
// same cloud provider account without conflict.
(define Namespace "REPLACED_IN_TEST_RUN")

// Defines the set of addresses that are allowed to access Quilt VMs.
(define AdminACL (list "local"))

(define MasterCount 1)
(define WorkerCount 1)
(docker "google/pause")
(label "red"  (makeList WorkerCount       (docker "google/pause")))
(label "blue" (makeList (* 3 WorkerCount) (docker "google/pause")))
(connect (list 1024 65535) "red" "blue")
(connect (list 1024 65535) "blue" "red")
