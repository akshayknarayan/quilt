(import "haproxy")
(import "memcached")
(import "mysql")
(import "spark")
(import "wordpress")
(import "zookeeper")

(define (link spark db)
  (if (and spark db)
    (progn
      (connect 7077 (hmapGet spark "master") (hmapGet db "slave"))
      (connect 7077 (hmapGet spark "worker") (hmapGet db "slave")))))

(define (New zone nCache nSql nWordpress nHaproxy nSparkM nSparkW nZoo)
  (let ((memcd (memcached.New (+ zone "-memcd") nCache))
        (db (mysql.New (+ zone "-mysql") nSql))
        (wp (wordpress.New (+ zone "-wp") nWordpress db memcd))
        (hap (haproxy.New (+ zone "-hap") nHaproxy wp))
        (zk (zookeeper.New (+ zone "-zk") nZoo))
        (spk (spark.New (+ zone "-spk") nSparkM nSparkW zk)))
    (link spk db)
    hap))
